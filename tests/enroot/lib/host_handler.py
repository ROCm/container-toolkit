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

import paramiko
import subprocess
import time
import re
import logging
import json
from scp import SCPClient

log = logging.getLogger(__name__)

class RemoteHostHandler:
    """
    This Class creates handle to the remote host to execute commands on the host
    """
    def __init__(self, host_ip):
        self.host_ip = host_ip
        self.client =  paramiko.SSHClient()
        self.output = {}
        self.sftp = ""

    def connect(self, username, password, key):
        """
            This method performs connect to the Node
    
            Parameters:
                username: string 
                password: string
      
            Returns:
                SSH Client object
        """
        self.username = username
        self.password = password
        self.key = key
        try:
            log.info(f"SSH to Node: {self.host_ip} using password")
            self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            if not password :
                self.client.connect(hostname=self.host_ip, username=username, key_filename=key)
            else:
                self.client.connect(hostname=self.host_ip, username=username, password=password)

        except Exception as e:
            log.error(f"Failed SSH connection to Device {self.host_ip}")
            log.exception(e)
        else:
            self.log.info(f"Connected Successfully to Device {self.host_ip}")

        finally:
            return self.client

    def execute_command(self, command):
        """
           This method executes the given command on the node
           Parameters: 
              command : command to execute on the device
           Returns:
              exit_code : int 
              output[] : output[] object having output['stdout'],output['stderr'],output['stdin']

        """
        try :
            
            log.info(f"Command to be executed on {self.host_ip}: {command} ")
            self.output['stdin'], self.output['stdout'], self.output['stderr'] = self.client.exec_command(command)
            exit_code = self.output['stdout'].channel.recv_exit_status()
            self.output['stdout'] = self.output['stdout'].read().decode()
            self.output['stderr'] = self.output['stderr'].read().decode()

        except Exception as e:
            log.error(f"Command failed : {command} on the Device: {self.host_ip}")
            log.exception(e)
            exit_code = 1
        
        return exit_code, self.output

    def execute_command_channel(self,command):
        """
        """
        try :
            log.info(f"Command to be executed on {self.host_ip} : {command} ")
            channel = self.client.get_transport().open_session()
            channel.get_pty() 
            channel.exec_command(command)


        except Exception as e:
            log.error(f"Command failed : {command} on the Device: {self.host_ip}")
            log.exception(e)
            exit_code = 1

        finally:
            if channel:
                log.info(f"Closing the channel")
                channel.close()
        return 0

    def copy_to_host(self,localpath,remotepath):
        """
            This method copies the file from local host to the remote host 
        """
        try:
            ftp_client = self.client.open_sftp()
            ftp_client.put(localpath,remotepath)
            ftp_client.close()
        except Exception as e:
            log.error(f"Unable to copy the package to host Error: {e}")
            return 1
        else:
            log.info(f"Copied {localpath} to {remotepath} successfully !")
            return 0

    def copy_from_host(self,remotepath,localpath):
        """
            This method copies the file from remote host to the local host 
        """
        try:
            ftp_client = self.client.open_sftp()
            ftp_client.stat(remotepath) 
            ftp_client.get(remotepath,localpath)
            ftp_client.close()
        except FileNotFoundError as f:
            log.error(f"{remotepath} file not found : {f}")
            return 1
        except Exception as e:
            log.error(f"Unable to copy the package from host Error: {e}")
            return 1

        else:
            log.info(f"Copied {remotepath} to {localpath} successfully !")
            return 0

    def create_file(self, remote_file_path, remote_file_content, is_json=False):
        """
        This method creates file on the remote host
        """
        try:
            if self.sftp == "":
                self.sftp = self.client.open_sftp()
            with self.sftp.open(remote_file_path, 'w') as f:
                if is_json :
                    json.dump(remote_file_content,f,indent=4)
                else :
                    f.write(remote_file_content)
        except Exception as e:
            log.error(f"Unable to create the file : {e}")
            return 1
        else:
            log.info(f"Created the file {remote_file_path} on {self.host_ip} successfully ! ")
            return 0
        finally :
            if self.sftp:
                self.sftp.close()
                self.sftp = ""
        
    def open_file(self,remote_file_path):
        """
        This method opens file on the remote host
        """
        try:
            self.sftp = self.client.open_sftp()
            with self.sftp.open(remote_file_path, 'r') as f:
                file_config = json.load(f)
        except Exception as e:
            log.error(f"Unable to open the file : {e}")
            return 1, e
        else:
            log.info(f"Opened the file {remote_file_path} on {self.host_ip} successfully ! ")
            return 0, file_config

    def copy_munge_to_hosts(self,dest_hosts, file_path):

        try:
            if self.sftp == "":
                self.sftp = self.client.open_sftp()
                self.sftp.get(file_path, "munge.key")
            
            dest_path = "munge.key"
            for host in dest_hosts:
                log.info(f"Copying {file_path} to {host.host_ip}...")
                scp = SCPClient(host.client.get_transport())
                scp.put("munge.key", dest_path)
                scp.close()

        except Exception as e:
            log.error(f"Unable to copy the file : {e}")
            return 1
        else:
            log.info(f"Copied the file {file_path} on all hosts successfully !")
            return 0
        finally :
            if self.sftp:
                self.sftp.close()
                self.sftp = ""
        
    def get_ip(self):
        """
        This method retrieves the IP Address of the remote host
        """
        try :
            transport = self.client.get_transport()
            peername = transport.getpeername()
            ip_address = peername[0]
        except Exception as e:
            return 1,e
        return 0, ip_address

    def close(self):
        """
            This method closes the connection from the Node
        """
        log.info(f"Attempting to Disconnect from Node: {self.host_ip}")
        try:
            self.client.close()
        except Exception as e:
            log.error(f"Unable to gracefully disconnect from Device: {self.host_ip} - (Socked Expired/ Closed already)")
        else:
            log.info(f"Closed SSH Connection Successfully for device: {self.host_ip}")
    
    def reconnect(self):
        """
        """
        self.close()
        self.connect(self.username,self.password,self.key)