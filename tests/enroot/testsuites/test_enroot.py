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
import re
import pytest
import time
from lib.helper_lib import HelperLib
from lib.host_handler import RemoteHostHandler
from utils import *
from pathlib import Path

repo_root = Path(__file__).resolve().parent.parent
config_folder = repo_root/"config"
batch_scripts_folder = repo_root/"batch_scripts"
helper_scripts_folder = repo_root/"helper_scripts"

log = logging.getLogger(__name__)

@pytest.fixture(scope='session', autouse=True)
def setup_and_teardown():
    setup_test()
    yield
    teardown_test()

@pytest.mark.level1
@pytest.mark.level2
def setup_test():
    """
    Install slurm, enroot, pyxis and setup the testbed

    TestID: TCID-ENROOT-SETUP

    Setup:
        1. Install slurm
        2. Install enroot
        3. Install Pyxis
    Validation:
        1. Verify if slurm, enroot and pyxis are correctly installed
    Raises:
        AssertionError: Above validation points are failed
    """

    # Check rocm version 
    log.info("Getting rocm version installed on the host..")
    for amd_host in  pytest.testdata.amd_host:
        exit_code , output = amd_host.helper_obj.get_rocmsmi_version()
        if exit_code :
            log.error(f"Rocm version couldnt be determined, Error : {output}")
            assert False, f"Rocm Version couldnt be determined, Error : {output}"
        else :
            log.info(f"Rocm Version is {output}")

    # Run rocm-smi
    for amd_host in  pytest.testdata.amd_host:
        log.info(f"Listing the GPUs on the host {amd_host.host_ip} using rocm-smi")
        exit_code, output = amd_host.execute_command(f"sudo rocm-smi")
        if exit_code :
            assert False , f" rocm-smi command execution failed !! , {output['stderr']}"
        log.debug(f"{output['stdout']}")
        amd_host.gpu_info = parse_rocm_smi_result(output['stdout'])
        amd_host.gpu_num = len(amd_host.gpu_info)
        exit_code, output = amd_host.execute_command(f"sudo rocm-smi --showuniqueid ")
        if exit_code :
            assert False , f" rocm-smi command execution failed !! , {output['stderr']}"
        log.debug(f"GPU info : {amd_host.gpu_info}, pytest.testdata.gpu_num : {amd_host.gpu_num} ")
        log.info(f"Total number of AMD GPUS on the device : {amd_host.gpu_num}")

    if pytest.no_install:
        log.info("Setup installation skipped... ")
        return

    # Uninstall slurm
    uninstall_script = "uninstall_slurm.sh"
    local_uninstall_script = config_folder/uninstall_script    
    for amd_host in  pytest.testdata.amd_host:
        log.info(f"Uninstalling slurm on {amd_host.host_ip}... ")
        amd_host.helper_obj.run_scripts(local_uninstall_script, uninstall_script,pytest.testdata.results_dir)
        log.info(f"Uninstalling slurm on {amd_host.host_ip}... SUCCESSFUL  !!")

    exit_code, output = pytest.testdata.amd_host[0].execute_command(f"sudo hostname -s ")
    if exit_code :
        assert False , f" Failed to get the host name !! , {output['stderr']}"  
    head_node = output['stdout'].strip()

    # Create config files
    node_name=[]
    for amd_host in  pytest.testdata.amd_host:
        exit_code,output = get_node_name(amd_host)
        if exit_code:
            assert False , f" Failed to getting the Node name !! , {output['stderr']}"  
        node_name.append(output)
        log.info(f"{node_name}")
    
    # Installation and config file creation
    for amd_host in  pytest.testdata.amd_host:
        # Create /etc/slurm/slurm.conf
        log.info(f"Creating /etc/slurm/slurm.conf on {amd_host.host_ip}...")
        local_slurm_conf = config_folder / "slurm.conf"
        exit_code = create_conf_file(amd_host,local_slurm_conf,head_node,node_name)
        if exit_code:
            assert False, f"/etc/slurm/slurm.conf couldnt be created!!"
        log.info(f"Creating /etc/slurm/slurm.conf on {amd_host.host_ip} - Successfull !!")

        # Create /etc/slurm/gres.conf
        log.info(f"Creating /etc/slurm/gres.conf on {amd_host.host_ip}...")
        exit_code = create_gres_conf_file(amd_host)
        if exit_code:
            assert False, f"/etc/slurm/gres.conf couldnt be created!!"
        log.info(f"Creating /etc/slurm/gres.conf  on {amd_host.host_ip}- Successfull !!")

        # Create /etc/slurm/cgroup.conf
        log.info(f"Creating /etc/slurm/cgroup.conf on {amd_host.host_ip}...")
        exit_code = create_cgroup_conf_file(amd_host)
        if exit_code:
            assert False, f"/etc/slurm/cgroup.conf couldnt be created!!"
        log.info(f"Creating /etc/slurm/cgroup.conf  on {amd_host.host_ip} - Successfull !!")

        # Add the user to render/video groups 
        exit_code, output = amd_host.execute_command(f"whoami ")
        if exit_code :
            assert False , f" Failed to get the user name !! , {output['stderr']}"  
        user_name = output['stdout'].strip() 
        log.info(f"Adding {user_name} to groups render/video on {amd_host.host_ip} ...")
        exit_code, output = amd_host.execute_command(f"sudo usermod -aG render,video {user_name}")
        if exit_code :
            assert False , f" Failed to add the user to render,video groups !! , {output['stderr']}" 
        log.info(f"Adding {user_name} to groups render/video on {amd_host.host_ip} Successfull !!")
        # Reconnecting the host handle after adding the user to render,video groups
        amd_host.reconnect()

        # Install slurm 
        install_script = "install_slurm.sh"
        local_install_script = config_folder/install_script
        log.info(f"Installing slurm on {amd_host.host_ip}... ")
        amd_host.helper_obj.run_scripts(local_install_script,install_script,pytest.testdata.results_dir,pytest.testdata.slurm_version)
        log.info(f"Installing slurm on {amd_host.host_ip}... SUCCESSFUL  !!")

        # Install  enroot 
        install_enroot = "install_enroot.sh"
        local_install_enroot = config_folder/install_enroot
        log.info(f"Installing enroot on {amd_host.host_ip} ...")
        amd_host.helper_obj.run_scripts(local_install_enroot,install_enroot,pytest.testdata.results_dir,pytest.testdata.enroot_version)
        log.info(f"Installing enroot on {amd_host.host_ip} ... SUCCESSFUL  !!")

        log.info(f"Setup complete on {amd_host.host_ip}..")

    # # Configure /etc/hosts file 
    host_entries=[]
    for amd_host in  pytest.testdata.amd_host:
        if not pytest.testdata.slurm_ip:
            exit_code, ip_address = amd_host.get_ip()
            if exit_code :
                assert False, f"Could not retrieve the remote server's IP Address !!"
            exit_code, output = amd_host.execute_command(f"sudo hostname -s ")
        else:
            ip_address = pytest.testdata.slurm_ip
        if exit_code :
            assert False , f" Failed to get the host name !! , {output['stderr']}"  
        host_name = output['stdout'].strip()

        host_entries.append(f"{ip_address} {host_name}")

    hosts_file = "/etc/hosts"
    for amd_host in  pytest.testdata.amd_host:
        for entry in host_entries:
            #log.info(f"Adding---{entry}--- to {amd_host.host_ip} /etc/hosts file... ")
            command = f"grep -qF \"{entry}\" {hosts_file} || echo \"{entry}\" | sudo tee -a {hosts_file} > /dev/null"
            exit_code, output = amd_host.execute_command(command)
            if exit_code :
                assert False , f" Failed to update the {hosts_file} !! , {output['stderr']}"  
            log.info(f"Adding---{entry}--- to {amd_host.host_ip} /etc/hosts file-Successfull !! ")

    # Create /etc.munge/munge.key and change file permission 
    exit_code, output = pytest.testdata.amd_host[0].helper_obj.create_munge_key()
    if exit_code :
        assert False , f"Munge key creation on {pytest.testdata.amd_host[0].host_ip} failed :{output['stderr']} "  
    log.info(f"Munge key creation on {pytest.testdata.amd_host[0].host_ip} successful ")
    
    # Copy to all the hosts
    munge_path = "/etc/munge/munge.key"
    exit_code = pytest.testdata.amd_host[0].copy_munge_to_hosts(pytest.testdata.amd_host[1:],munge_path )
    if exit_code:
        assert False, "Munge key copy to all the hosts failed !!"
    
    # Change back the permission of all munge keys to 700 and restart munge,slurm and slurmctld
    exit_code, output = pytest.testdata.amd_host[0].helper_obj.configure_head_node()
    if exit_code :
        assert False, f"Head node configuration on {pytest.testdata.amd_host[0].host_ip} failed :{output['stderr']} "  
    log.info(f"Head node configuration on  {pytest.testdata.amd_host[0].host_ip} successfull ")

    for amd_host in  pytest.testdata.amd_host[1:]:
        exit_code, output = amd_host.helper_obj.configure_munge()
        if exit_code :
            assert False, f"Munge key configuration on {amd_host.host_ip} failed :{output['stderr']} "  
        log.info(f"Munge key configuration on {amd_host.host_ip} successfull ")
    
    # Configuring slurmdb
    amd_host=pytest.testdata.amd_host[0]
    local_slurmdbd_file = config_folder / "slurmdbd.conf"
    log.info(f"Creating /etc/slurm/slurmdbd.conf on {amd_host.host_ip}...")
    exit_code = create_conf_file(amd_host,local_slurmdbd_file )
    if exit_code:
        assert False, f"/etc/slurm/slurmdbd.conf couldnt be created!!"
    log.info(f"Creating /etc/slurm/slurmdbd.conf  on {amd_host.host_ip} - Successfull !!")
    
    log.info(f"Configuring Slurmdbd ...")
    slurmdb_config = "slurmdb_config.sh"
    local_slurmdb_config = config_folder/slurmdb_config
    amd_host.helper_obj.run_scripts(local_slurmdb_config,slurmdb_config,pytest.testdata.results_dir)  
    log.info(f"Slurmdbd configuration on {amd_host.host_ip} successfull ")
    
    for amd_host in  pytest.testdata.amd_host:
        exit_code, output = amd_host.execute_command("sudo systemctl restart slurmd")
        if exit_code :
            assert False, f"slurmd restart failed on {amd_host.host_ip} : {output['stderr']}"
        log.info(f"slurmd restart on {amd_host.host_ip} : \n {output['stdout']}")

    amd_host = pytest.testdata.amd_host[0]
    exit_code, output = amd_host.execute_command("sudo sacctmgr list cluster")
    if exit_code :
        assert False, f"failed to get sacct cluster on {amd_host.host_ip} : {output['stderr']}"
    log.info(f"sacct cluster  : \n {output['stdout']}")    
    amd_host.execute_command("sudo systemctl restart slurmctld")

    install_pyxis = "install_pyxis.sh"
    local_install_pyxis = config_folder/install_pyxis
    for amd_host in  pytest.testdata.amd_host:
        # Install  pyxis 
        log.info(f"Installing pyxis on {amd_host.host_ip} ...")
        amd_host.helper_obj.run_scripts(local_install_pyxis,install_pyxis,pytest.testdata.results_dir)
        log.info(f"Installing pyxis on {amd_host.host_ip} ... SUCCESSFUL  !!")

    for amd_host in  pytest.testdata.amd_host:
        exit_code, output = amd_host.execute_command("sinfo")
        if exit_code :
            assert False, f"sinfo failed on {amd_host.host_ip} : {output['stderr']}"
        log.info(f"sinfo on {amd_host.host_ip} : \n {output['stdout']}")
        
def test_single_node_pytorch():
    """    
    Use sbatch to run a single node pytorch test

    TestID: TCID-ENROOT-SINGLE-PYTORCH

    Setup:
        1. Create temporary folder /tmp/test_pytorch
        2. Copy gpu stress script to this folder
        3. Run sbatch script 
    Validation:
        1. Verify if sbatch test is completed
        2. Verify and output the results 
    Raises:
        AssertionError: Above validation points are failed
    """

    # Create batch script
    head_node = pytest.testdata.amd_host[0]
    local_script = batch_scripts_folder / "pytorch_gpu_util_sbatch.sh"
    remote_script = str(local_script.name)
    log.info(f"Creating {local_script.name} on {head_node.host_ip}...")
    exit_code = create_batch_script(head_node,local_script)
    if exit_code:
        assert False, f"{local_script.name} on {head_node.host_ip} couldnt be created!!"
    log.info(f"Creating {local_script.name} on {head_node.host_ip} - Successfull !!")

    for amd_host in  pytest.testdata.amd_host:
        # Create /tmp/test_pytorch/gpu_stress_10s.py
        parent_dir = "/tmp/test_pytorch"
        copy_file_list =[]
        local_stress_script = helper_scripts_folder / "gpu_stress_10s.py"
        log.info(f"Creating {local_stress_script.name} on {amd_host.host_ip}...")
        exit_code = create_helper_script(amd_host,local_stress_script,parent_dir)
        if exit_code:
            assert False, f"{local_stress_script.name} on {amd_host.host_ip} couldnt be created!!"
        log.info(f"Creating {local_stress_script.name} on {amd_host.host_ip} - Successfull !!")

        #Get host name 
        exit_code, output = amd_host.execute_command(f"sudo hostname -s ")
        if exit_code :
            assert False , f" Failed to get the host name !!, {output['stderr']}"  
        node = output['stdout'].strip()

        # Run the batch script -> get jobid 
        sbatch_cmd = f"sbatch --parsable --nodelist={node}  --gres=gpu:{amd_host.gpu_num} {remote_script} "
        exit_code, output = head_node.execute_command(sbatch_cmd)
        assert not exit_code, f"sbatch command couldnt be launched !! : {output['stderr']}"
        job_id = output['stdout'].strip()
        log.info(f"sbatch job - {job_id} submitted !!") 

        # Wait for job completion
        job_state, sacct_output = wait_for_job_completion(head_node,job_id) 
        log.info(f"Job state of {job_id} : {job_state}")
        log.info(f"sacct output : {sacct_output}")
        err_file = f"pytorch_logs/pytorch-util-{job_id}.err"
        output_file = f"pytorch_logs/pytorch-util-{job_id}.out"
        copy_file_list.append(output_file)
        copy_file_list.append(err_file)
        
        if "COMPLETED" not in job_state:
            exit_code, output = amd_host.execute_command(f"cat {err_file}")
            assert not exit_code, f"{amd_host.host_ip}:Couldnt print the batch error file {err_file} : {output['stderr']}"
            log.info(f"ERROR file : {output['stdout']}")
            assert False, "Pytorch_gpu_util test case failed.. !! "

        # Print rocm-smi, cuda device_count output
        exit_code, output = amd_host.execute_command(f"cat {output_file} | head -n 20")
        assert not exit_code, f"{amd_host.host_ip}:Couldnt print the batch output file {output_file} : {output['stderr']}"
        log.info(f"Output : ")
        log.info(output['stdout'].encode().decode('unicode_escape'))

        # Check for gpu_max_utilization.log in /tmp/test_pytorch and validate
        log.info(f"Checking {parent_dir}/gpu_max_utilization.log ...")
        gpu_util_log = f"{parent_dir}/gpu_max_utilization.log"
        copy_file_list.append(gpu_util_log)
        exit_code, output = amd_host.execute_command(f"cat {parent_dir}/gpu_max_utilization.log ")
        if exit_code :
            assert False , f" Error retrieving the file {parent_dir}/gpu_max_utilization.log !, {output['stderr']}"  
        log.info(f"Output : ")
        log.info(output['stdout'].encode().decode('unicode_escape'))
        # Copy back results and deleted the directory and files
        log.info(f"Copying all the results to {str(pytest.testdata.results_dir)}...")
        
        for file in copy_file_list:
            local_file = pytest.testdata.results_dir / Path(file).name
            exit_code = amd_host.copy_from_host(file,local_file)
            assert not exit_code, f" Error copying the file {file} !"
            exit_code, output = amd_host.execute_command(f"sudo rm -rf {file}")
            if exit_code :
                assert False , f" Error deleting the file {file} !, {output['stderr']}"  

        # Remove the parent directory
        exit_code, output = amd_host.execute_command(f"sudo rm -rf {parent_dir}")
        if exit_code :
            assert False , f" Error deleting the folder {parent_dir} !, {output['stderr']}"  

    # Delete the batch script on the remote host 
    exit_code, output = head_node.execute_command(f"sudo rm -rf {remote_script}")
    if exit_code :
        assert False , f" Error deleting the script {remote_script}!, {output['stderr']}"  

def test_multi_node_distributed_pytorch():
    """    
    Use sbatch to run distributed pytorch test on multiple nodes

    TestID: TCID-ENROOT-MULTI-NODE-PYTORCH

    Setup:
        1.Create test_pytorch dir  
        2.Copy batch file to the home directory
        3.Copy helper script - distributed_pytorch.py to the test_pytorch directory
    Validation:
        1. Verify if sbatch test is completed
        2. Verify if the output file - test-summary-{job-id}.txt is created and print that output
        2. Verify and print the results
    Raises:
        AssertionError: Above validation points are failed
    """
    # Create helper script
    parent_dir = "test_pytorch"
    copy_file_list =[]
    for amd_host in  pytest.testdata.amd_host:
        local_pytorch_script = helper_scripts_folder / "distributed_pytorch.py"
        log.info(f"Creating {local_pytorch_script.name} on {amd_host.host_ip}...")
        exit_code = create_helper_script(amd_host,local_pytorch_script,parent_dir)
        if exit_code:
            assert False, f"{local_pytorch_script.name} on {amd_host.host_ip} couldnt be created!!"
        log.info(f"Creating {local_pytorch_script.name} on {amd_host.host_ip} - Successfull !!")

    amd_host = pytest.testdata.amd_host[0]
    # Create batch script
    local_script = batch_scripts_folder / "distributed_pytorch_sbatch.sh"
    remote_script = str(local_script.name)
    log.info(f"Creating {local_script.name} on {amd_host.host_ip}...")
    exit_code = create_batch_script(amd_host,local_script)
    if exit_code:
        assert False, f"{local_script.name} on {amd_host.host_ip} couldnt be created!!"
    log.info(f"Creating {local_script.name} on {amd_host.host_ip} - Successfull !!")
    
    # Run the batch script -> get jobid 
    exit_code, output = amd_host.execute_command(f"sbatch --parsable --gres=gpu:{amd_host.gpu_num} {remote_script} ")
    assert not exit_code, f"sbatch command couldnt be launched !! : {output['stderr']}"
    job_id = output['stdout'].strip()
    log.info(f"sbatch job - {job_id} submitted !!")  

    # Wait for job completion
    job_state, sacct_output = wait_for_job_completion(amd_host,job_id) 
    log.info(f"Job state of {job_id} : {job_state}")
    log.info(f"sacct output : {sacct_output}")
    err_file = f"pytorch_logs/pytorch-rccl-{job_id}.err"
    output_file = f"pytorch_logs/pytorch-rccl-{job_id}.out"
    copy_file_list.append(output_file)
    copy_file_list.append(err_file)

    if "COMPLETED" not in job_state:
        exit_code, output = amd_host.execute_command(f"cat {err_file}")
        assert not exit_code, f"{amd_host.host_ip}:Couldnt print the batch error file {err_file} : {output['stderr']}"
        log.info(f"ERROR file : {output['stdout']}")
        assert False, "Distributed Pytorch test case failed.. !! "

    # Check for test_summary.txt in test_pytorch and validate
    log.info(f"Checking {parent_dir}/ ...")
    test_summary_log = f"{parent_dir}/test_summary_{job_id}.txt"
    copy_file_list.append(test_summary_log)
    exit_code, output = amd_host.execute_command(f"cat {test_summary_log} ")
    assert not exit_code, f" Error retrieving the file {test_summary_log}!, {output['stderr']}"  
    log.info(f"Output : ")
    log.info(output['stdout'].encode().decode('unicode_escape'))
 
    # Copy back results and delete the directory and files
    log.info(f"Copying all the results to {str(pytest.testdata.results_dir)}...")
    
    for file in copy_file_list:
        local_file = pytest.testdata.results_dir / Path(file).name
        exit_code = amd_host.copy_from_host(file,local_file)
        assert not exit_code, f" Error copying the file {file} !"
        exit_code, output = amd_host.execute_command(f"sudo rm -rf {file}")
        assert not exit_code , f" Error deleting the file {file} !, {output['stderr']}"  

    # Remove the parent directory
    exit_code, output = amd_host.execute_command(f"sudo rm -rf {parent_dir}")
    assert not exit_code, f" Error deleting the folder {parent_dir} !, {output['stderr']}"  

    # Delete the batch script on the remote host 
    exit_code, output = amd_host.execute_command(f"sudo rm -rf {remote_script}")
    assert not exit_code , f" Error deleting the script {remote_script}!, {output['stderr']}"  

def  test_multi_node_rccl():
    """    
    Use sbatch to run rccl test on multiple nodes

    TestID: TCID-ENROOT-MULTI-NODE-RCCL

    Setup:
        1.Copy batch file to the home directory
        2.Launch the sbatch script

    Validation:
        1. Verify if sbatch test is completed
        2. Verify if the output file - logs/rccl_test_%j.out is created and print that output
        2. Verify and print the results
    Raises:
        AssertionError: Above validation points are failed
    """

    amd_host = pytest.testdata.amd_host[0]
    copy_file_list =[]
    # Create batch script
    local_script = batch_scripts_folder / "rccl_tests_sbatch.sh"
    remote_script = str(local_script.name)
    log.info(f"Creating {local_script.name} on {amd_host.host_ip}...")
    exit_code = create_batch_script(amd_host,local_script)
    if exit_code:
        assert False, f"{local_script.name} on {amd_host.host_ip} couldnt be created!!"
    log.info(f"Creating {local_script.name} on {amd_host.host_ip} - Successfull !!")
    
    # Run the batch script -> get jobid 
    exit_code, output = amd_host.execute_command(f"sbatch --parsable --gres=gpu:{amd_host.gpu_num} {remote_script} ")
    assert not exit_code, f"sbatch command couldnt be launched !! : {output['stderr']}"
    job_id = output['stdout'].strip()
    log.info(f"sbatch job - {job_id} submitted !!")  

    # Wait for job completion
    job_state, sacct_output = wait_for_job_completion(amd_host,job_id) 
    log.info(f"Job state of {job_id} : {job_state}")
    log.info(f"sacct output : {sacct_output}")
    err_file = f"logs/rccl_test_{job_id}.err"
    output_file = f"logs/rccl_test_{job_id}.out"
    copy_file_list.append(output_file)
    copy_file_list.append(err_file)

    if "COMPLETED" not in job_state:
        exit_code, output = amd_host.execute_command(f"cat {err_file}")
        assert not exit_code, f"{amd_host.host_ip}:Couldnt print the batch error file {err_file} : {output['stderr']}"
        log.info(f"ERROR file : {output['stdout']}")
        assert False, "RCCL test case failed.. !! "

    # Check for output file and print the results
    parent_dir="logs"
    log.info(f"Checking {parent_dir}/ ...")
    exit_code, output = amd_host.execute_command(f"cat {output_file} ")
    assert not exit_code, f" Error retrieving the file {output_file}!, {output['stderr']}"  
    log.info(f"Output : ")
    log.info(output['stdout'].encode().decode('unicode_escape'))
 
    # Copy back results and delete the directory and files
    log.info(f"Copying all the results to {str(pytest.testdata.results_dir)}...")
    
    for file in copy_file_list:
        local_file = pytest.testdata.results_dir / Path(file).name
        exit_code = amd_host.copy_from_host(file,local_file)
        assert not exit_code, f" Error copying the file {file} !"
        exit_code, output = amd_host.execute_command(f"sudo rm -rf {file}")
        assert not exit_code , f" Error deleting the file {file} !, {output['stderr']}"  

    # Remove the parent directory
    exit_code, output = amd_host.execute_command(f"sudo rm -rf {parent_dir}")
    assert not exit_code, f" Error deleting the folder {parent_dir} !, {output['stderr']}"  

    # Delete the batch script on the remote host 
    exit_code, output = amd_host.execute_command(f"sudo rm -rf {remote_script}")
    assert not exit_code , f" Error deleting the script {remote_script}!, {output['stderr']}"  

def teardown_test():
    """
    Teardown the testbed
    1.Uninstall slurm
    2.Uninstall enroot

    """
    if pytest.no_uninstall:
        log.info("Setup uninstallation skipped... Nothing to teardown...!")
        return
    uninstall_script = "uninstall_slurm.sh"
    local_uninstall_script = config_folder/uninstall_script   
    pytorch_logs = "pytorch_logs"  
    for amd_host in  pytest.testdata.amd_host:
        # Remove the logs directory
        log.info(f"Removing {pytorch_logs} directory")
        exit_code, output = amd_host.execute_command(f"sudo rm -rf {pytorch_logs}")
        assert not exit_code, f" Error deleting the folder {pytorch_logs} !, {output['stderr']}"  

        log.info(f"Uninstalling slurm on {amd_host.host_ip}... ")
        amd_host.helper_obj.run_scripts(local_uninstall_script, uninstall_script,pytest.testdata.results_dir)
        log.info(f"Uninstalling slurm on {amd_host.host_ip}... SUCCESSFUL  !!")

        log.info(f"Uninstalling enroot on {amd_host.host_ip}... ")
        exit_code, output = amd_host.execute_command(f"""yes "Y" | sudo  DEBIAN_FRONTEND=noninteractive apt purge enroot """)
        if exit_code :
            assert False , f" Error uninstalling enroot on {amd_host.host_ip}, {output['stderr']}"  
        
    log.info("Testbed teardown complete")    
