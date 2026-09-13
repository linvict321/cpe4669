import sys
import subprocess

# node_list = [1, 2, 4, 8, 16]
node_list = [8, 16]
dim_list = [100, 1000, 5000, 10000]

for nodes in node_list:
    for dim in dim_list:
        print(f"\nRunning matmul.py with {nodes} nodes and matrix dimension {dim}...")
        command = f"mpiexec -n {nodes} python ./matmul.py {dim}"
        result = subprocess.run(command, shell=True)
        if result.returncode != 0:
            print(f"Error: matmul.py failed with exit code {result.returncode}.")
            continue