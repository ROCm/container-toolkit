#!/usr/bin/env python3

# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the \"License\");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an \"AS IS\" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import torch
import torch.multiprocessing as mp
import threading
import subprocess
import time
import os
from typing import List, Tuple, Optional

LOG_FILE = "gpu_max_utilization.log"
UTIL_THRESHOLD = 95      # percent threshold considered "max"
WROTE_EVENT = threading.Event()


# --------------------------
# Worker: multi-stream GEMM
# --------------------------
def stress_gpu(gpu_id: int, duration: int = 30, num_streams: int = 8, N: int = 16384):
    # each process sets its own device
    torch.cuda.set_device(gpu_id)
    device = torch.device("cuda", gpu_id)

    # allocate once
    a = torch.randn(N, N, device=device)
    b = torch.randn(N, N, device=device)

    streams = [torch.cuda.Stream(device=device) for _ in range(num_streams)]
    print(f"[GPU {gpu_id}] Starting stress: duration={duration}s, streams={num_streams}, N={N}")

    end = time.time() + duration
    while time.time() < end:
        for s in streams:
            with torch.cuda.stream(s):
                _ = a @ b
        # ensure completion of all streams before next wave
        torch.cuda.synchronize()
    print(f"[GPU {gpu_id}] Stress finished.")


# --------------------------
# Parse rocm-smi CSV output
# --------------------------
def parse_rocm_smi_csv() -> Tuple[Optional[List[str]], Optional[List[List[str]]]]:
    try:
        out = subprocess.check_output(["rocm-smi", "--showuse", "--csv"], text=True, stderr=subprocess.DEVNULL)
    except FileNotFoundError:
        print("rocm-smi not found in PATH.")
        return None, None
    except subprocess.CalledProcessError as e:
        print(f"rocm-smi call failed: {e}")
        return None, None
    except Exception as e:
        print(f"rocm-smi error: {e}")
        return None, None

    lines = [ln.replace('"', "") for ln in out.strip().splitlines() if ln.strip()]
    if len(lines) < 2:
        return None, None

    header = [h.strip() for h in lines[0].split(",")]
    rows = [[cell.strip() for cell in r.split(",")] for r in lines[1:]]
    return header, rows


# --------------------------
# Build printable table
# --------------------------
def build_table_text(header: List[str], rows: List[List[str]]) -> str:
    cols = len(header)
    col_widths = [len(h) for h in header]
    for r in rows:
        for i in range(min(cols, len(r))):
            col_widths[i] = max(col_widths[i], len(str(r[i])))

    header_line = " | ".join(header[i].ljust(col_widths[i]) for i in range(cols))
    sep = "-+-".join("-" * col_widths[i] for i in range(cols))
    text = header_line + "\n" + sep + "\n"
    for r in rows:
        text += " | ".join((r[i] if i < len(r) else "").ljust(col_widths[i]) for i in range(cols)) + "\n"
    return text


# --------------------------
# Convert a utilization string to integer percent
# --------------------------
def parse_percent(s: str) -> Optional[int]:
    if s is None:
        return None
    s = s.strip()
    if s == "":
        return None
    # remove trailing percent signs or other text
    s = s.replace("%", "").replace(",", "").split()[0]
    try:
        val = int(float(s))
        return val
    except Exception:
        return None


# --------------------------
# Determine if all GPUs exceed threshold
# --------------------------
def all_gpus_at_threshold(header: List[str], rows: List[List[str]], threshold: int) -> bool:
    # find the first header column that contains "use" (case-insensitive)
    util_idx = None
    for i, h in enumerate(header):
        if "use" in h.lower():
            util_idx = i
            break
    if util_idx is None:
        # cannot find utilization column
        return False

    # parse each row's util column
    for r in rows:
        if util_idx >= len(r):
            return False
        p = parse_percent(r[util_idx])
        if p is None:
            return False
        if p < threshold:
            return False
    return True


# --------------------------
# Monitor thread (prints live, writes log once)
# --------------------------
def monitor_thread_func(stop_time: float, warmup: float = 3.0, poll_interval: float = 1.0):
    global WROTE_EVENT
    # small warmup to let workers start
    time.sleep(warmup)
    while time.time() < stop_time:
        header, rows = parse_rocm_smi_csv()
        if header is None or rows is None:
            # parsing failed — print a message and retry
            print("[monitor] rocm-smi returned no usable output, retrying...")
            time.sleep(poll_interval)
            continue

        # build and print table
        table_text = build_table_text(["timestamp"] + header, [[time.strftime("%H:%M:%S")] + r for r in rows])
        print("\033c", end="")  # clear (may not be supported everywhere)
        print(table_text)

        # check threshold
        if not WROTE_EVENT.is_set():
            # pass header and rows as-is (header has column names without timestamp)
            if all_gpus_at_threshold(header, rows, UTIL_THRESHOLD):
                try:
                    with open(LOG_FILE, "w") as f:
                        f.write(table_text)
                    print(f"\n*** Max utilization detected — snapshot saved to {LOG_FILE} ***")
                    WROTE_EVENT.set()
                except Exception as e:
                    print(f"[monitor] failed to write log: {e}")
        time.sleep(poll_interval)


# --------------------------
# Main
# --------------------------
def main():
    if not torch.cuda.is_available():
        print("CUDA/ROCm not available to PyTorch (torch.cuda.is_available() is False). Exiting.")
        return

    ngpus = torch.cuda.device_count()
    print(f"Detected {ngpus} GPUs")

    # clear previous log
    try:
        open(LOG_FILE, "w").close()
    except Exception:
        pass

    duration = 30
    stop_time = time.time() + duration

    mp.set_start_method("spawn", force=True)

    # spawn stress processes (one per GPU)
    procs = []
    for gpu in range(ngpus):
        p = mp.Process(target=stress_gpu, args=(gpu, duration))
        p.start()
        procs.append(p)

    # start monitor thread in main process
    monitor = threading.Thread(target=monitor_thread_func, args=(stop_time, 3.0, 1.0), daemon=True)
    monitor.start()

    # join workers
    for p in procs:
        p.join()

    # wait until monitor finishes (or until it wrote)
    monitor.join(timeout=5)

    if WROTE_EVENT.is_set():
        print(f"\nFinal GPU max-utilization snapshot saved → {LOG_FILE}")
    else:
        print(f"\n(No max utilization snapshot written.)")


if __name__ == "__main__":
    main()