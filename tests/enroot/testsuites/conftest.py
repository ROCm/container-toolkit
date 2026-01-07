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
import yaml
from datetime import datetime
from yaml.loader import SafeLoader
from  lib.host_handler import RemoteHostHandler
from lib.helper_lib import HelperLib
from utils import *
from pathlib import Path


log = logging.getLogger(__name__)

class testcache(object):
    amd_hosts = []
    helper_obj = {}
    amd_ctk_list = {}
    gpu_num = 0
    gpu_info = []
    host_type = {}
    testbed = {}
    results_dir = ""
    slurm_version = ""
    enroot_version = ""
    slurm_ip = ""

testdata = testcache()

def pytest_configure(config):
    pytest.testbed_dir = config.getoption("--testbed")
    pytest.no_install = config.getoption("--no-install")
    pytest.no_uninstall = config.getoption("--no-uninstall")
    testdata.results_dir = results_dir()
    config.option.log_file = str(testdata.results_dir / "pytest.log")

def pytest_addoption(parser):
    parser.addoption("--testbed", action="store", default=None, help="Testbed yaml file for remote host details")    
    parser.addoption("--no-install", action="store_true", help="Skip installation steps (enabled by default)")
    parser.addoption("--no-uninstall", action="store_true",help="Skip uninstallation steps (enabled by default)")
    
@pytest.fixture(scope="session", autouse=True)
def testbed_initialize():
    if pytest.testbed_dir:
        parse_inventory()
    connect_handles()
    
    pytest.testdata = testdata
    yield
    if pytest.testbed_dir:
        close_handles()

def parse_inventory():
    inventory_file = pytest.testbed_dir
    if not os.path.exists(inventory_file):
        log.info('Inventory File %s does not exist' % inventory_file, type='error')
    with open(inventory_file, 'r') as file:
        testdata.testbed = yaml.safe_load(file)
    
def connect_handles():
    amd_hosts = []
    for host in testdata.testbed.values():
        amd_host = RemoteHostHandler(host['host'])
        rsa_key = Path.home() / ".ssh" / "id_rsa"
        amd_host.connect(host['user'], host.get('password',None),key=host.get('key',str(rsa_key)))
        testdata.slurm_version = host.get('slurm_version',"")
        testdata.enroot_version = host.get('enroot_version',"")
        testdata.slurm_ip = host.get('slurm_ip',"")
        amd_host.helper_obj = HelperLib(amd_host)
        amd_hosts.append(amd_host)
    testdata.amd_host = amd_hosts
    
def close_handles():
    for amd_host in testdata.amd_hosts:
        amd_host.close()

def results_dir():
    timestamp = datetime.now().strftime("results-%Y-%m-%d_%H-%M-%S")
    results_dir = Path(__file__).resolve().parent.parent / "results"
    run_dir = results_dir / timestamp

    run_dir.mkdir(parents=True, exist_ok=True)

    return run_dir