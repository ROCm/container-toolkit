#!/usr/bin/env python3
import sys
import subprocess
import re
from pathlib import Path

def update_docker_image(test_name, docker_image):
    """Update DOCKER_IMAGE in the appropriate batch script."""
    if test_name == "test_single_node_pytorch":
        script_path = Path("batch_scripts/pytorch_gpu_util_sbatch.sh")
        pattern = r'DOCKER_IMAGE=.*'
        replacement = f'DOCKER_IMAGE={docker_image}'
        print(f"Updating pytorch_gpu_util_sbatch.sh with image: {docker_image}")
    elif test_name == "test_multi_node_distributed_pytorch":
        script_path = Path("batch_scripts/distributed_pytorch_sbatch.sh")
        pattern = r'export DOCKER_IMAGE=.*'
        replacement = f'export DOCKER_IMAGE={docker_image}'
        print(f"Updating distributed_pytorch_sbatch.sh with image: {docker_image}")
    else:
        print(f"Unknown test name: {test_name}")
        return
    
    content = script_path.read_text()
    updated_content = re.sub(pattern, replacement, content)
    script_path.write_text(updated_content)

def main():
    if len(sys.argv) < 2:
        print("Usage: run_test.py <test_name> [docker_image] [no_install] [no_uninstall]")
        sys.exit(1)
    
    test_name = sys.argv[1]
    docker_image = sys.argv[2] if len(sys.argv) > 2 and sys.argv[2] else None
    no_install = sys.argv[3] if len(sys.argv) > 3 else "false"
    no_uninstall = sys.argv[4] if len(sys.argv) > 4 else "false"
    
    # Update Docker image if provided
    if docker_image:
        update_docker_image(test_name, docker_image)
    
    # Build pytest command
    cmd = [
        "python3", "-m", "pytest",
        "testsuites/test_enroot.py",
        "--testbed", "testbed/enroot_tb.yml",
        "-k", test_name
    ]
    
    if no_install == "true":
        cmd.append("--no-install")
    
    if no_uninstall == "true":
        cmd.append("--no-uninstall")
    
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd)
    sys.exit(result.returncode)

if __name__ == "__main__":
    main()
