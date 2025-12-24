# Guide to run enroot tests

This guide explains how to set up your environment and run Python tests using **pytest**.

---

## Prerequisites

On the remote GPU host :

- GPU drivers should be installed and GPUS should be detected
- **Rocm** should be installed and rocm-smi should be working
- Make sure the /etc/hostname has the correct name of the device. 

Ensure the following are installed on your test runner node:

- **Python 3.8+**
- **pip** (Python package manager)
- (Optional but recommended) **virtualenv** or **venv**

Check versions:

```bash
python3 --version
pip3 --version
```

---

## Setup (Recommended: Virtual Environment)

Create and activate a virtual environment:

```bash
# Clone tests/enroot
cd enroot
python3 -m venv venv
source venv/bin/activate   # Linux / macOS
# venv\Scripts\activate    # Windows

# Add enroot directory to the python path 
export PYTHONPATH=/<home>/<user>/enroot/:$PYTHONPATH  # Linux
# $env:PYTHONPATH = "C:\Users\username\enroot\;" + $env:PYTHONPATH  # Windows
```

Upgrade pip and install dependencies

```bash
pip install --upgrade pip
pip install -r requirements.txt
```

---

## Update enroot_tb.yml

Before starting the test, provide server/node information in enroot_tb.yml:
```bash
host: # Mandatory : IP address of the GPU node 
user: # Mandatory : Username of the GPU node to be used for ssh 
password: # Optional if key is provided : Password for ssh access of the node
key: # Optional if password is provided:  Path to the ssh key
slurm_version: # Optional:  Version of slurm to be installed on the host , this key can be commented out if latest version is to be used. (Recomended to use same version on all hosts)
enroot_version: # Optional: Enroot version to be installed on the host, this key can be commented out if latest enroot version is to be used. (Recomended to use same version on all hosts)
```
For ssh authentication if password is to be used, provide password in single quotes.
Provide slurm and enroot version if needed. 
```bash
# Sample testbed yaml file 
host1:
  host: 11.22.33.44
  user: 'enroot'
  password: 'password'
  key: 'Path/to/the/key'
  slurm_version: '24.05.4'
  enroot_version: '4.0.1'
```
If key has to be used, provide the path to the key in single quotes and comment out the password line.  

```bash
# Sample testbed yaml file 
host1:
  host: 11.22.33.44
  user: 'enroot'
  #password: 'password'
  key: 'Path/to/the/key'
```
If there are multiple hosts, provide all the necessary details as follows.  

```bash
# Sample testbed yaml file 
host1:
  host: 11.22.33.44
  user: 'enroot'
  key: 'Path/to/the/key'

host2:
  host: 55.66.77.88
  user: 'enroot'
  key: 'Path/to/the/key'
```

## Running Tests

The script by default installs slurm,enroot and pyxis on the nodes and uninstalls them once the test is complete. 
All the logs and results are copied back to the **results** folder 

Test flow :
1. Testbed setup:
    * Check how many GPUs are available using "rocm-smi"
    * Install slurm, enroot and Pyxis(skip this if *--no-install* flag is given in the command line)
2. Run the test: 
    * Launch sbatch to run the test
    * Once the test is complete, copy back all the results and logs to "results" folder
3. Testbed teardown:
    a. Uninstall slurm, enroot and pyxis(skip this if *--no-uninstall* flag is given in the command line)

```bash
cd testsuites
python3 -m pytest test_enroot.py --testbed ../testbed/enroot_tb.yml
```

Run a specific test:

```bash
python3 -m pytest test_enroot.py --testbed ../testbed/enroot_tb.yml -k test_single_node_pytorch
```

Run a test and skip testbed cleanup at the end 

```bash
python3 -m pytest test_enroot.py --testbed ../testbed/enroot_tb.yml -k test_single_node_pytorch --no-uninstall
```
Run only the test and skip installation, if slurm, enroot and pyxis are already installed

```bash
python3 -m pytest test_enroot.py --testbed ../testbed/enroot_tb.yml -k test_single_node_pytorch --no-install
```

---
