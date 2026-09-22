#!/bin/bash
#SBATCH --job-name="mergesortmpi"
#SBATCH --output="mergesort.%j.%N.out"
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks-per-node=5
#SBATCH --mem=8G
#SBATCH --account="slo106"
#SBATCH --export=ALL
#SBATCH --time=00:10:00

#having issues with SLURM export in venv
export SLURM_EXPORT_ENV=ALL

#load modules
module purge
module load cpu/0.17.3b
module load gcc/10.2.0/npcyll4
module load openmpi/4.1.1
module load slurm

#preloads slurm globals into venv to get rid of auth errors
export LD_PRELOAD=/cm/shared/apps/slurm/23.02.7/lib64/libslurm.so

#virtual environment
source .venv/bin/activate

#Run programs in parallel
srun -n 1 python mpi_sort.py 0
srun -n 3 python mpi_sort.py 17
srun -n 5 python mpi_sort.py 100003