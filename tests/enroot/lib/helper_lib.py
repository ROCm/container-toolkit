#!/usr/bin/env python

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
import os
import tenacity
from pathlib import Path


log = logging.getLogger(__name__)

class HelperLib():

    def __init__(self, tbnode):
        self._tbnode = tbnode

    @property
    def host(self) -> str:
        return self._tbnode

    def get_hosttype(self):
        """
        This function gets the host type whether Ubuntu22 or Ubunut24 or RHEL

        Args : None

        Return : (int,string/[])
            0 : if the host type is either Ubuntu22 or Ubunut24 or RHEL
            1 : if the host type is neither Ubuntu22 or Ubunut24 or RHEL
        """
        exit_code, result = self.host.execute_command("cat /etc/os-release")
        
        if exit_code == 0:
            log.debug(f"{result['stdout']}")
            match = re.search(r"PRETTY_NAME=\"(.*)\"", result["stdout"])
            if match :
                log.info(f"Host OS is : {match.group(1)}")
                if "Ubuntu 22" in match.group(1):
                    host_type = "Ubuntu22"
                    return 0, host_type
                elif "Ubuntu 24" in match.group(1):
                    host_type = "Ubuntu24"
                    return 0, host_type
                elif "Rocky Linux" in match.group(1):
                    host_type = "RHEL"
                    return 0, host_type
                else :
                    log.info(f"This OS is not supported ! OS : {match.group(1)} ")
                    return 1 , result
        else :
            log.info(f"Unable to check the OS Version. Error : {result['stderr']}")
            return exit_code, result
        
    def get_rocmsmi_version(self):
        """
        This function gets the rocm version installed on the host

        Args : None 

        Return: int, string
        """
        exit_code, result = self.host.execute_command("sudo update-alternatives --display rocm")
        if exit_code == 0 :
            match = re.search(r"link currently points to \/opt\/rocm-(\d+\.\d+)(.*)", result['stdout'])
            if match:
                rocm_version = match.group(1)
                if float(rocm_version) < 6.4 :
                    result['stderr'] = f"Rocm-smi version ({rocm_version}) is lesser than 6.4 ! Please use rocm-verson 6.4 or above for amd-ctk usage !"
                    return 1, result['stderr']
                else :
                    return 0, rocm_version
            else :
                return 2, f"Rocm version couldnt be found in {result['stdout']}"
        else :
            return exit_code , result['stderr']

    def run_scripts(self,local_script,remote_script, results_dir,version=None):
        """
        """
        log_file= remote_script.replace(".sh", f"_{self.host.host_ip}.log")
        # Copy script to the host
        exit_code = self.host.copy_to_host(local_script,remote_script)
        if exit_code:
            log.info(f"Could not copy {local_script}")
            return exit_code
        # Give exec permission to the script
        exit_code, output = self.host.execute_command(f"chmod +x {remote_script}")
        if exit_code:
            log.err(f"Error giving exec permission to {remote_script} : {output['stderr']}")
            return exit_code,output
        self.host.execute_command_channel(f"sudo nohup ./{remote_script} {version} > {log_file} ")
        self.wait_for_script_completion(remote_script,log_file)

        # Copy log from the host
        local_log_file  = results_dir / log_file
        exit_code = self.host.copy_from_host(log_file,local_log_file)
        if exit_code:
            log.info(f"Could not copy {local_script}")
            return exit_code

        # Delete the log file and script file on remote host 
        exit_code, output = self.host.execute_command(f"sudo rm -rf {log_file} {remote_script}")
        if exit_code:
            log.err(f"Error deleting the log: {log_file} : {output['stderr']}")
            return exit_code
        log.info("Deleted the script and the log from the host...")
        
    @tenacity.retry(wait=tenacity.wait_fixed(15), stop=tenacity.stop_after_attempt(150))
    def wait_for_script_completion(self,script,log_file):
        log.info(f"Checking {log_file}...")
        exit_code, output = self.host.execute_command(f"cat {log_file} ")
        exit_code, output = self.host.execute_command(f"ps -ef | grep {script} | grep -v grep  ")
        if f"{script}" in output['stdout']:
            raise Exception(f'{script} is still running... ')
        return True

    def create_munge_key(self):
        """
        """
        exit_code , result = self.get_hosttype()
        if exit_code == 0:
            if result == "Ubuntu22" or result == "Ubuntu24":
                commands = [
                "sudo rm -f /etc/munge/munge.key ",
                "sudo -u munge /usr/sbin/mungekey --verbose",
                "sudo chmod 777 -R  /etc/munge",
                "sudo chown -R munge: /etc/munge/munge.key"
                ]
        else :
            return exit_code, result

        for cmd in commands:
            exit_code, result = self.host.execute_command(f"{cmd}") 
            if exit_code:
                return exit_code, result
        
        return 0, "munge_create_done"

    def configure_munge(self):
        """
        """
        exit_code , result = self.get_hosttype()
        if exit_code == 0:
            if result == "Ubuntu22" or result == "Ubuntu24":
                commands = [
                    "sudo cp munge.key /etc/munge/munge.key",
                    "sudo chmod 700 /etc/munge/munge.key",
                    "sudo chown -R munge: /etc/munge/munge.key",
                    "sudo  systemctl enable munge",
                    "sudo systemctl restart munge",
                    "sudo systemctl restart slurmd"
                ]
        else :
            return exit_code, result

        for cmd in commands:
            exit_code, result = self.host.execute_command(f"{cmd}") 
            if exit_code:
                return exit_code, result
        
        return 0, "munge_config_done"
    
    def configure_head_node(self):
        """
        """
        exit_code , result = self.get_hosttype()
        if exit_code == 0:
            if result == "Ubuntu22" or result == "Ubuntu24":
                commands = [
                    "sudo chmod 700 -R /etc/munge",
                    "sudo chown -R munge: /etc/munge/munge.key",
                    "sudo systemctl enable munge",
                    "sudo systemctl restart munge",
                    "sudo systemctl restart slurmctld",
                    "sudo systemctl restart slurmd"
                ]
        else :
            return exit_code, result

        for cmd in commands:
            exit_code, result = self.host.execute_command(f"{cmd}") 
            if exit_code:
                return exit_code, result
        
        return 0, "headnode_config_done"



