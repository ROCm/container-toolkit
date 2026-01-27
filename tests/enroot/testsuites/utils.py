#! /usr/bin/env python3
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

import logging
import pytest
import os
import re
import yaml
import tenacity
import textwrap
from pathlib import Path

from lib.host_handler import RemoteHostHandler
from lib.helper_lib import HelperLib

log = logging.getLogger(__name__)

def parse_rocm_smi_result(output):
        gpu_info = []
        lines = output.splitlines()
        for line in lines:
            # Skip header or irrelevant lines
            if line.startswith('Device') or 'DID' in line or '===' in line:
                continue
            if 'Ignoring' in line  or line.startswith('WARNING')  or "allocated" in line or "invalid" in line :
                continue
            
            # Extract GPU details 
            columns = line.split()
            if len(columns) >= 3:  
                gpu_data = {
                    'node_id':columns[1],
                    'gpu_id': columns[3],
                    'NPS_type': columns[6],
                    'compute_type':columns[7],
                    'partition_id' : columns[8],
                    'Usage': columns[15],
                }
                gpu_info.append(gpu_data)

        return gpu_info
 
def parse_test_output(output, expected_gpu_num):
    """
    Return:
        True : if correct number of GPUS are attached
        False : if incorrect number of GPUs are attached or rocm-smi output is incorrect
    """
    match = re.search(r"= ROCm System Management Interface =", output)
    if match :
        log.debug(f"The command output : {output}")
        test_gpu_info = parse_rocm_smi_result(output)
        gpu_num_result = len(test_gpu_info)
        log.info(f"Expected GPU num : {expected_gpu_num}    Attached GPU: {gpu_num_result}")
        if gpu_num_result == expected_gpu_num: 
            log.info(f"Expected and attached GPU numbers are same: {gpu_num_result}")
            return True        
        else:
            log.info(f"Expected and attached GPUs are not same")
            return False
    else:
        log.info("Rocm-smi output couldn't be parsed !")
        return False

def is_hexadecimal(s):
    """
    Checks if a given string represents a valid hexadecimal number.
    Args:
        s: The string to check.
    Returns:
        True if the string is a valid hexadecimal number, False otherwise.
    """
    if not isinstance(s, str):
        return False  # Ensure input is a string
    if not s:
        return False  # Empty string is not a valid hex number
    try:
        int(s, 10)
        return False
    except ValueError:
        int(s, 16)
        return True

def create_conf_file(amd_host,local_file,head_node=None,node_name=None):
    """
    Create /etc/slurm/slurm.conf file on the remote host 
    """
    # Create file contents
    tmp_file ="/tmp/" + local_file.name
    etc_file =  "/etc/slurm/" + local_file.name
    
    # Copy the conf file to the host
    exit_code = amd_host.copy_to_host(local_file,tmp_file)
    if exit_code:
        log.info(f"Could not copy {local_file}")
        return exit_code
    if local_file.name == "slurm.conf":
        multi_node = "\\\n".join(node_name)
        command = f"sudo sed -i \
                   -e 's/HEAD-NODE/{head_node}/g' \
                   -e 's|MULTI-NODE|{multi_node}|g' \
                   {tmp_file}"
        exit_code, output = amd_host.execute_command(command)
        if exit_code:
            log.info(f"Error modifying {tmp_file} : {output['stderr']}")
            return exit_code

    exit_code, output = amd_host.execute_command(f"sudo mkdir -p /etc/slurm && sudo cp {tmp_file} {etc_file}")
    if exit_code:
        log.info(f"Error copying {tmp_file} to {etc_file} folder : {output['stderr']}")
        return exit_code
    return exit_code

def get_node_name(amd_host):
    """
    Run a script on the host to get the output similar to slurmd -C output 
    """
    script_name = "get_node_name.sh"
    get_node_name_script="""
    #!/bin/bash

    NODENAME=$(hostname -s)
    CPUS=$(lscpu | grep '^CPU(s):' | awk '{print $2}')
    SOCKETS=$(lscpu | grep '^Socket(s):' | awk '{print $2}')
    CORES_PER_SOCKET=$(lscpu | grep '^Core(s) per socket:' | awk '{print $4}')
    THREADS_PER_CORE=$(lscpu | grep '^Thread(s) per core:' | awk '{print $4}')

    # Get total physical memory in KB, then convert to MB (Slurm uses MB)
    # We subtract a little (e.g. 500MB or more for large systems) for OS overhead
    TOTAL_MEM_KB=$(free -k | grep Mem: | awk '{print $2}')
    REAL_MEM_MB=$(( ($TOTAL_MEM_KB / 1024) - 512 ))

    echo "NodeName=${NODENAME} CPUs=${CPUS} Boards=1 SocketsPerBoard=${SOCKETS} CoresPerSocket=${CORES_PER_SOCKET} ThreadsPerCore=${THREADS_PER_CORE} RealMemory=${REAL_MEM_MB}"
    """
    output = ""
    exit_code = amd_host.create_file(script_name,get_node_name_script)
    if exit_code:
        return exit_code,output
    exit_code, output = amd_host.execute_command(f"sudo sh {script_name}")
    if exit_code:
        log.info(f"Error getting the output for nodename : {output['stderr']}")
        return exit_code,output
    node_info = output['stdout'].strip()
    exit_code, output = amd_host.execute_command(f"sudo rm -rf  {script_name}")
    if exit_code:
        log.info(f"Error deleting the script : {output['stderr']}")
        return exit_code,output
    return 0, node_info  

def create_gres_conf_file(amd_host):
    """
    Create /etc/slurm/gres.conf file on the remote host 
    """
    # Create file contents
    tmp_gres_file ="/tmp/gres.conf"
    etc_gres_file =  "/etc/slurm/gres.conf"
    gres_conf_content = textwrap.dedent("""
        AutoDetect=rsmi
        """).lstrip()

    exit_code = amd_host.create_file(tmp_gres_file,gres_conf_content)
    if exit_code:
        log.info(f"Could not create {tmp_gres_file}")
        return exit_code
    exit_code, output = amd_host.execute_command(f"sudo cp {tmp_gres_file} {etc_gres_file}")
    if exit_code:
        log.info(f"Error copying gres.conf to /etc/slurm folder : {output['stderr']}")
        return exit_code
    return exit_code

def create_cgroup_conf_file(amd_host):
    """
    Create /etc/slurm/cgroup.conf file on the remote host 
    """
    # Create file contents
    tmp_cgroup_file ="/tmp/cgroup.conf"
    etc_cgroup_file =  "/etc/slurm/cgroup.conf"
    cgroup_conf_content = textwrap.dedent("""
        ConstrainCores=yes
        ConstrainDevices=yes
        ConstrainRAMSpace=yes
        """).lstrip()
    exit_code = amd_host.create_file(tmp_cgroup_file,cgroup_conf_content)
    if exit_code:
        log.info(f"Could not create {tmp_cgroup_file}")
        return exit_code
    exit_code, output = amd_host.execute_command(f"sudo cp {tmp_cgroup_file} {etc_cgroup_file}")
    if exit_code:
        log.info(f"Error copying cgroup.conf to /etc/slurm folder : {output['stderr']}")
        return exit_code
    return exit_code
    
def create_helper_script(amd_host,local_helper_script,parent_dir="/tmp/test_pytorch" ):
    """
    Create /tmp/test_pytorch dir and /tmp/test_pytorch/gpu_stress_10s.py file on the remote host 
    """
    # Create file contents
    remote_helper_script =  f"{parent_dir}/" + local_helper_script.name
    
    exit_code, output = amd_host.execute_command(f"mkdir -p {parent_dir}")
    if exit_code:
        log.info(f"Error creating {parent_dir} : {output['stderr']}")
        return exit_code

    exit_code = amd_host.copy_to_host(local_helper_script,remote_helper_script)
    if exit_code:
        log.info(f"Could not copy {local_helper_script}")
        return exit_code
    

    exit_code, output = amd_host.execute_command(f"chmod +x {remote_helper_script}")
    if exit_code:
        log.info(f"Error giving execute permission to {remote_helper_script}: {output['stderr']}")
        return exit_code

    return exit_code

def create_batch_script(amd_host,local_batch_script):
    """
    Create batch script file on the remote host 
    """
    remote_batch_script = str(local_batch_script.name)
    exit_code = amd_host.copy_to_host(local_batch_script,remote_batch_script)
    if exit_code:
        log.info(f"Could not copy {local_batch_script}")
        return exit_code

    
    exit_code, output = amd_host.execute_command(f"chmod +x {remote_batch_script}")
    if exit_code:
        log.info(f"Error giving execute permission to {remote_batch_script}: {output['stderr']}")
        return exit_code
    return exit_code

@tenacity.retry(wait=tenacity.wait_fixed(20), stop=tenacity.stop_after_attempt(60))
def wait_for_job_completion(headnode, job_id):

    # Get job status from sacct
    exit_code, output = headnode.execute_command(f"sacct -j {job_id} --format=JobID,State --noheader")
    if exit_code:
        log.info(f"Error getting the sacct job status : {output['stderr']}")
        return exit_code, output
    #Parse sacct output 
    sacct_output  = output['stdout']
    lines = [l for l in sacct_output.splitlines() if l.strip()]
    if not lines:
        log.info("No sacct result yet, retrying...")
        raise Exception("Job state not available yet")

    cols = lines[0].split()
    state = cols[1] if len(cols) > 1 else None
    
    log.info(f"Current job state: {state}")
    if state in ["COMPLETED", "FAILED", "CANCELLED", "TIMEOUT", "OUT_OF_MEMORY"]:
        return state, sacct_output

    raise Exception(f"Job still running: {state}")

def list_all_ib_devices_remote(amd_host):
    exit_code, output = amd_host.execute_command(f"ls /sys/class/infiniband")
    if exit_code:
        log.error(f"Error listing IB devices: {output['stderr']}")
        return exit_code,0

    devices = []
    for dev in output['stdout'].split():
        ports_cmd = f"ls /sys/class/infiniband/{dev}/ports"
        exit_code, output = amd_host.execute_command(ports_cmd)
        if exit_code:
            log.error(f"Error listing IB devices Ports: {output['stderr']}")
            return exit_code,0
        ports_out = output['stdout']
        for port in ports_out.split():
            devices.append((dev, int(port)))

    if not devices:
        return exit_code,0

    return exit_code, devices

def read_ib_counters_remote(amd_host, device, port):
    base = f"/sys/class/infiniband/{device}/ports/{port}/hw_counters"

    counters = {}
    for name in (
        "tx_rdma_ucast_bytes",
        "rx_rdma_ucast_bytes",
        "tx_rdma_ucast_pkts",
        "rx_rdma_ucast_pkts",
    ):
        cmd = f"cat {base}/{name} 2>/dev/null || echo 0"
        exit_code, output = amd_host.execute_command(cmd)
        if exit_code:
            log.error(f"Error listing IB devices: {output['stderr']}")
            return exit_code,0
        val = output['stdout']
        counters[name] = int(val)

    return counters

def parse_used_ib_devices_from_log(log_path):

    NET_IB_REGEX = re.compile(
    r'\[\d+\]([a-zA-Z0-9_]+):(\d+)/(?:RoCE|IB)',
    re.IGNORECASE)

    devices = set()
    net_ib_lines = []
    socket_lines = []

    with open(log_path, "r", errors="ignore") as f:
        for line in f:
            line = line.strip()

            if "NET/Socket" in line:
                socket_lines.append(line)

            if "NET/IB" not in line:
                continue

            net_ib_lines.append(line)

            for m in NET_IB_REGEX.finditer(line):
                dev, port = m.groups()
                devices.add((dev, int(port)))

    if not net_ib_lines:
        raise AssertionError("No NET/IB lines found in NCCL log")

    if socket_lines:
        raise AssertionError(
            "Socket fallback detected:\n" +
            "\n".join(socket_lines[:3])
        )

    return sorted(devices), net_ib_lines

def counter_delta(before, after):
    return {k: after[k] - before[k] for k in before}

