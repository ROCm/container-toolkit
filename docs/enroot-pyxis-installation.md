# Enroot and Pyxis 

Enroot and Pyxis are tools created to run containerized AI/HPC workloads on SLURM. These tools can now be used on a SLURM cluster  with AMD GPUs to make them run efficiently and achieve isolation for these  GPUs. Traditional runtimes like Docker/Podman bring additional overhead such as daemons, root privileges and extra storage layers. With Enroot, users can convert Docker images into a simple unpacked filesystem tree and run containers as a regular Linux process. Also, with Enroot and Pyxis, each job is granted exclusive GPU device files which prevents jobs from accidentally accessing the same GPU device.  

Enroot reference: https://github.com/NVIDIA/enroot 

Pyxis reference:  https://github.com/NVIDIA/pyxis 

This guide provides the steps to install enroot/pyxis on a SLURM cluster as well as examples to run containerized images isolating specific AMD GPUs on Ubuntu. 

Installation: 

Pre-requisites: 
Make sure SLURM is already installed  and the cluster is up and running.  
Since GPUs are used with enroot and pyxis, the /etc/slurm/gres.conf should be  configured with the correct renderD IDs.  

Sample gres.conf file : 
```bash
Name=gpu Type=rocm  File=/dev/dri/renderD128 
Name=gpu Type=rocm  File=/dev/dri/renderD136 
Name=gpu Type=rocm  File=/dev/dri/renderD144 
Name=gpu Type=rocm  File=/dev/dri/renderD152 
Name=gpu Type=rocm  File=/dev/dri/renderD160 
Name=gpu Type=rocm  File=/dev/dri/renderD168 
Name=gpu Type=rocm  File=/dev/dri/renderD176 
Name=gpu Type=rocm  File=/dev/dri/renderD184 
```
Enroot Installation: 
```bash
#Check requirements 
curl -fSsL -O https://github.com/NVIDIA/enroot/releases/download/v3.5.0/enroot-check_3.5.0_$(uname -m).run 
chmod +x enroot-check_*.run 
./enroot-check_*.run --verify 
./enroot-check_*.run 
```

Install enroot through a package 
```bash
arch=$(dpkg --print-architecture) 
curl -fSsL -O https://github.com/NVIDIA/enroot/releases/download/v3.5.0/enroot_3.5.0-1_${arch}.deb 
curl -fSsL -O https://github.com/NVIDIA/enroot/releases/download/v3.5.0/enroot+caps_3.5.0-1_${arch}.deb 
sudo apt install -y ./*.deb 
```
Validate Enroot installation  
```bash
enroot import docker://rocm/pytorch:latest 
enroot create rocm+pytorch+latest.sqsh 
enroot start rocm+pytorch+latest rocm-smi 
ENROOT_RESTRICT_DEV=y enroot start rocm+pytorch+latest rocm-smi 
```
Reference: https://github.com/NVIDIA/enroot/blob/master/doc/installation.md 

## Steps to build and install Pyxis  
Install the following packages 

```bash
sudo apt update 
sudo apt install -y devscripts 
sudo apt install -y debhelper 
```
Create a deb package 
```bash
git clone https://github.com/NVIDIA/pyxis 
cd pyxis 
git checkout v0.20.0 
make orig 
CPPFLAGS="-I/usr/local/slurm-24.05.5.1/include" LDFLAGS="-L/usr/local/slurm-24.05.5.1/lib" make deb 
```
After this step, nvslurm-plugin-pyxis_0.20.0-1_amd64.deb will be created in the same directory. 

## Steps to install pyxis deb on all the compute nodes and also the slurm head-node 
Install the same pyxis deb package on the headnode and all the compute nodes. 
While installing pyxis on the headnode/controller node, it will throw error that enroot is not installed but we can ignore this error since we need not have enroot on the head-bode.  

```bash
sudo dpkg -i ./nvslurm-plugin-pyxis_0.20.0-1_amd64.deb 
sudo mkdir /etc/slurm/plugstack.conf.d 
sudo ln -s /usr/share/pyxis/pyxis.conf /etc/slurm/plugstack.conf.d/pyxis.conf 
sudo touch /etc/slurm/plugstack.conf 
echo "include /etc/slurm/plugstack.conf.d/*" | sudo tee -a /etc/slurm/plugstack.conf 
```

Restart slurmd on the compute  
```bash
sudo systemctl restart slurmd 
```
Restart slurmd and slurmctld on all nodes 
```bash
sudo systemctl restart slurmd 
sudo systemctl restart slurmctld 
```
### Test with SLURM 

Following shows 4 isolated AMD GPUs running a containerized image rocm/pytorch 

```bash
ubuntu@node-4:~$ srun --gres=gpu:4 --container-image=docker://rocm/pytorch:latest rocm-smi 
pyxis: importing docker image: docker://rocm/pytorch:latest 
pyxis: imported docker image: docker://rocm/pytorch:latest 
============================================ ROCm System Management Interface ============================================ 

====================================================== Concise Info ====================================================== 

Device  Node  IDs              Temp        Power     Partitions          SCLK    MCLK    Fan  Perf  PwrCap  VRAM%  GPU% 

              (DID,     GUID)  (Junction)  (Socket)  (Mem, Compute, ID) 

========================================================================================================================== 

0       2     0x74a1,   28851  45.0Â°C      137.0W    NPS1, SPX, 0        123Mhz  900Mhz  0%   auto  750.0W  0%     0% 

1       3     0x74a1,   43178  41.0Â°C      133.0W    NPS1, SPX, 0        124Mhz  900Mhz  0%   auto  750.0W  0%     0% 

2       4     0x74a1,   32898  44.0Â°C      133.0W    NPS1, SPX, 0        124Mhz  900Mhz  0%   auto  750.0W  0%     0% 

3       5     0x74a1,   22683  40.0Â°C      136.0W    NPS1, SPX, 0        124Mhz  900Mhz  0%   auto  750.0W  0%     0% 

========================================================================================================================== 

================================================== End of ROCm SMI Log =================================================== 
```

Following shows isolating 2 AMD GPUs running a containerized image rocm/pytorch 

```bash
ubuntu@node-4:~$ srun --gres=gpu:2 --container-image=docker://rocm/pytorch:latest rocm-smi 
pyxis: importing docker image: docker://rocm/pytorch:latest 
pyxis: imported docker image: docker://rocm/pytorch:latest 
============================================ ROCm System Management Interface ============================================ 

====================================================== Concise Info ====================================================== 

Device  Node  IDs              Temp        Power     Partitions          SCLK    MCLK    Fan  Perf  PwrCap  VRAM%  GPU% 

              (DID,     GUID)  (Junction)  (Socket)  (Mem, Compute, ID) 

========================================================================================================================== 

0       2     0x74a1,   28851  45.0Â°C      137.0W    NPS1, SPX, 0        123Mhz  900Mhz  0%   auto  750.0W  0%     0% 

1       3     0x74a1,   43178  41.0Â°C      133.0W    NPS1, SPX, 0        124Mhz  900Mhz  0%   auto  750.0W  0%     0% 

========================================================================================================================== 

================================================== End of ROCm SMI Log =================================================== 
```

Following command runs a test.py script to different torch.cuda variables 

```bash
ubuntu@node-4:~$ srun --gres=gpu:2 --container-image=docker://rocm/pytorch:latest --container-mounts="$HOME:/home/$MY_USER" python3 /home/$MY_USER/test.py 
pyxis: importing docker image: docker://rocm/pytorch:latest 
pyxis: imported docker image: docker://rocm/pytorch:latest 
--- PyTorch CUDA Status --- 
torch.cuda.is_available(): True 
torch.cuda.device_count(): 2 

Visible Devices (from CUDA_VISIBLE_DEVICES env var): 
  CUDA_VISIBLE_DEVICES=0,1 
Detected Devices: 
  Device 0: AMD Instinct MI300X 
    Capability: (9, 4) 
    Memory (GB): 191.98 
  Device 1: AMD Instinct MI300X 
    Capability: (9, 4) 
    Memory (GB): 191.98 
--- End PyTorch CUDA Status --- 

The following command can be used to save the image locally to use for subsequent runs 
ubuntu@node-4:~$ srun --gres=gpu:8 --container-image=docker://rocm/pytorch:latest --container-save=/var/lib/ubuntu/enroot/rocm+pytorch+latest.sqsh rocm-smi 
ubuntu@node-4:~$ srun --gres=gpu:8 --container-image=/var/lib/ubuntu/enroot/rocm+pytorch+latest.sqsh rocm-smi 

Following command shows the usage of --exclusive directive with 2 GPUS requested.  
root@head-node:~# srun --exclusive --gres=gpu:2 --container-image=./rocm+ubuntu.sqsh --pty bash 
root@valid-prawn:/# rocm-smi 
============================================ ROCm System Management Interface ============================================ 

====================================================== Concise Info ====================================================== 

Device  Node  IDs              Temp        Power     Partitions          SCLK    MCLK    Fan  Perf  PwrCap  VRAM%  GPU%   

              (DID,     GUID)  (Junction)  (Socket)  (Mem, Compute, ID)                                                   

========================================================================================================================== 

0       2     0x74a1,   28851  35.0Â°C      142.0W    NPS1, SPX, 0        132Mhz  900Mhz  0%   auto  750.0W  0%     0%     

1       3     0x74a1,   44463  34.0Â°C      136.0W    NPS1, SPX, 0        134Mhz  900Mhz  0%   auto  750.0W  0%     0%     

========================================================================================================================== 

================================================== End of ROCm SMI Log =================================================== 
```
