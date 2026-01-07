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
# limitations under the License

import torch
import torch.distributed as dist
import os
import sys
import socket
import datetime
import traceback

def log(message, rank=None):
    timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
    hostname = socket.gethostname().split('.')[0]
    rank_str = f"[Rank {rank}]" if rank is not None else "[INIT]"
    print(f"{timestamp} | {hostname} | {rank_str} | {message}", flush=True)

def main():
    try:
        log("="*80)
        log("PYTORCH DISTRIBUTED TEST - ROCM/RCCL")
        log("="*80)
        
        # Get configuration from environment
        master_addr = os.environ.get('MASTER_ADDR', 'localhost')
        master_port = os.environ.get('MASTER_PORT', '29500')
        world_size = int(os.environ.get('WORLD_SIZE', '1'))
        rank = int(os.environ.get('RANK', '0'))
        local_rank = int(os.environ.get('LOCAL_RANK', '0'))
        
        log(f"Master: {master_addr}:{master_port}")
        log(f"World size: {world_size}")
        log(f"Rank: {rank} (local: {local_rank})")
        
        # Track device info and test results
        device_type = "CPU"
        device_name = "CPU"
        device_index = -1
        backend_name = "gloo"
        allreduce_passed = False
        broadcast_passed = False
        allgather_passed = False
        
        # Set up device BEFORE initializing process group
        if torch.cuda.is_available():
            device = torch.device(f'cuda:{local_rank}')
            torch.cuda.set_device(device)
            backend = 'nccl'  # Use NCCL (RCCL on ROCm)
            backend_name = 'nccl/rccl'
            device_type = "GPU"
            device_name = torch.cuda.get_device_name(device)
            device_index = local_rank
            log(f"Using device: {device}", rank)
            log(f"GPU: {device_name}", rank)
            log(f"Using backend: {backend} (RCCL on ROCm)", rank)
        else:
            device = torch.device('cpu')
            backend = 'gloo'
            backend_name = 'gloo'
            log("Using CPU", rank)
            log(f"Using backend: {backend}", rank)
        
        log("="*80)
        log("INITIALIZING PROCESS GROUP")
        
        # Initialize with device_id to suppress warning
        init_kwargs = {
            'backend': backend,
            'init_method': 'env://',
            'world_size': world_size,
            'rank': rank,
            'timeout': datetime.timedelta(seconds=300)
        }
        
        # Add device_id for NCCL backend
        if backend == 'nccl':
            init_kwargs['device_id'] = device
            
        dist.init_process_group(**init_kwargs)
        log("✓ Process group initialized!", rank)
        log("="*80, rank)
        
        # TEST 1: ALLREDUCE
        log("="*80, rank)
        log("TEST 1: ALLREDUCE", rank)
        tensor = torch.ones(2, 2).to(device) * (rank + 1)
        log(f"Before allreduce:\n{tensor}", rank)
        
        dist.all_reduce(tensor, op=dist.ReduceOp.SUM)
        log(f"After allreduce:\n{tensor}", rank)
        
        expected = sum(range(1, world_size + 1))
        if torch.allclose(tensor, torch.ones(2, 2).to(device) * expected):
            log("✓ AllReduce PASSED", rank)
            allreduce_passed = True
        else:
            log(f"✗ AllReduce FAILED (expected {expected})", rank)
        
        # TEST 2: BROADCAST
        log("="*80, rank)
        log("TEST 2: BROADCAST", rank)
        
        if rank == 0:
            broadcast_tensor = torch.tensor([1.0, 2.0, 3.0, 4.0]).to(device)
            log(f"Rank 0 broadcasting: {broadcast_tensor}", rank)
        else:
            broadcast_tensor = torch.zeros(4).to(device)
            log(f"Before broadcast: {broadcast_tensor}", rank)
        
        dist.broadcast(broadcast_tensor, src=0)
        log(f"After broadcast: {broadcast_tensor}", rank)
        
        expected_bcast = torch.tensor([1.0, 2.0, 3.0, 4.0]).to(device)
        if torch.allclose(broadcast_tensor, expected_bcast):
            log("✓ Broadcast PASSED", rank)
            broadcast_passed = True
        else:
            log("✗ Broadcast FAILED", rank)
        
        # TEST 3: ALLGATHER
        log("="*80, rank)
        log("TEST 3: ALLGATHER", rank)
        
        local_tensor = torch.tensor([float(rank)], dtype=torch.float32).to(device)
        gathered = [torch.zeros(1, dtype=torch.float32).to(device) for _ in range(world_size)]
        
        dist.all_gather(gathered, local_tensor)
        
        log(f"Gathered tensors: {[t.item() for t in gathered]}", rank)
        
        expected_gathered = list(range(world_size))
        if [int(t.item()) for t in gathered] == expected_gathered:
            log("✓ AllGather PASSED", rank)
            allgather_passed = True
        else:
            log("✗ AllGather FAILED", rank)
        
        # Synchronize before summary
        dist.barrier()
        
        # Give time for buffered output to flush
        import time
        time.sleep(0.2 * rank)
        
        # Per-rank summary payload
        rank_info = {
            "rank": rank,
            "local_rank": local_rank,
            "hostname": socket.gethostname().split('.')[0],  # Short hostname
            "device_type": device_type,
            "device_name": device_name,
            "device_index": device_index,
            "backend": backend_name,
            "allreduce_passed": allreduce_passed,
            "broadcast_passed": broadcast_passed,
            "allgather_passed": allgather_passed,
        }
        
        # Gather all results on rank 0
        all_rank_info = [None] * world_size
        dist.all_gather_object(all_rank_info, rank_info)
        
        # Extra barrier and delay before writing summary
        dist.barrier()
        
        # Rank 0 writes consolidated summary to file AND prints to stdout
        if rank == 0:
            time.sleep(0.5)  # Let all other output settle
            
            # Determine output file path
            job_id = os.environ.get('SLURM_JOB_ID', 'unknown')
            summary_file = f"test_summary_{job_id}.txt"
            
            # Build summary content
            summary_lines = []
            summary_lines.append("=" * 110)
            summary_lines.append("CONSOLIDATED TEST SUMMARY".center(110))
            summary_lines.append("=" * 110)
            summary_lines.append(f"Job ID: {job_id}")
            summary_lines.append(f"Timestamp: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
            summary_lines.append(f"World Size: {world_size}")
            summary_lines.append(f"Backend: {backend_name}")
            summary_lines.append("")
            
            # Header
            header = (
                f"{'Rank':<6} {'Local':<6} {'Hostname':<20} "
                f"{'Device':<8} {'GPU Idx':<8} {'Device Name':<25} {'Backend':<12} "
                f"{'AllReduce':<10} {'Broadcast':<10} {'AllGather':<10}"
            )
            summary_lines.append(header)
            summary_lines.append("-" * 110)
            
            overall_allreduce = True
            overall_broadcast = True
            overall_allgather = True
            gpu_count = 0
            cpu_count = 0
            
            for info in sorted(all_rank_info, key=lambda x: x["rank"]):
                overall_allreduce &= info["allreduce_passed"]
                overall_broadcast &= info["broadcast_passed"]
                overall_allgather &= info["allgather_passed"]
                
                if info["device_type"] == "GPU":
                    gpu_count += 1
                else:
                    cpu_count += 1
                
                gpu_idx_str = str(info["device_index"]) if info["device_index"] >= 0 else "N/A"
                allreduce_status = "PASS" if info["allreduce_passed"] else "FAIL"
                broadcast_status = "PASS" if info["broadcast_passed"] else "FAIL"
                allgather_status = "PASS" if info["allgather_passed"] else "FAIL"
                
                row = (
                    f"{info['rank']:<6} {info['local_rank']:<6} {info['hostname']:<20} "
                    f"{info['device_type']:<8} {gpu_idx_str:<8} {info['device_name']:<25} {info['backend']:<12} "
                    f"{allreduce_status:<10} {broadcast_status:<10} {allgather_status:<10}"
                )
                summary_lines.append(row)
            
            summary_lines.append("-" * 110)
            summary_lines.append(f"Total GPUs used: {gpu_count}")
            summary_lines.append(f"Total CPUs used: {cpu_count}")
            summary_lines.append("")
            summary_lines.append("Test Results:")
            summary_lines.append(f"  AllReduce:  {'✓ PASSED' if overall_allreduce else '✗ FAILED'}")
            summary_lines.append(f"  Broadcast:  {'✓ PASSED' if overall_broadcast else '✗ FAILED'}")
            summary_lines.append(f"  AllGather:  {'✓ PASSED' if overall_allgather else '✗ FAILED'}")
            summary_lines.append("")
            
            if overall_allreduce and overall_broadcast and overall_allgather:
                summary_lines.append("🎉 ALL TESTS PASSED 🎉".center(110))
            else:
                summary_lines.append("⚠️  SOME TESTS FAILED ⚠️".center(110))
            
            summary_lines.append("=" * 110)
            
            # Write to file
            try:
                with open(summary_file, 'w') as f:
                    f.write('\n'.join(summary_lines))
                    f.write('\n')
                print(f"\n✓ Summary written to: {summary_file}", flush=True)
            except Exception as e:
                print(f"\n✗ Failed to write summary file: {e}", flush=True)
            
            # Also print to stdout
            print("\n" + '\n'.join(summary_lines) + "\n", flush=True)
        
        log("="*80, rank)
        log("✓ Rank completed, destroying process group", rank)
        
        dist.destroy_process_group()
        
    except Exception as e:
        log(f"ERROR: {e}")
        log(traceback.format_exc())
        sys.exit(1)

if __name__ == "__main__":
    main()

