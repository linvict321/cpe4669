import sys
import subprocess

"""Run with:
python nbody_tester.py
"""

NUM_BODIES = 1000
NUM_STEPS = 100

# first compile the sequential and octree/barnes-hut C programs
seq_compile_command = "gcc -O2 nbody.c -o nbody -lm"
subprocess.run(seq_compile_command, shell=True)
dist_compile_command = "mpicc -O2 -o distoct distoct.c"
subprocess.run(dist_compile_command, shell=True)

print(f"Running all tests with {NUM_BODIES} bodies for {NUM_STEPS} steps:\n")

seq_run_command = f"nbody {NUM_BODIES} {NUM_STEPS}"
print("Running Sequential implementation...")
subprocess.run(seq_run_command, shell=True)

ranks_list = [1, 2, 4, 8]

for ranks in ranks_list:
    print(f"\nRunning distoct with {ranks} ranks...")
    dist_run_command = f"mpiexec -np {ranks} distoct {NUM_BODIES} {NUM_STEPS}"
    result = subprocess.run(dist_run_command, shell=True)
    if result.returncode != 0:
        print(f"Error: distoct failed with exit code {result.returncode}.")
        continue