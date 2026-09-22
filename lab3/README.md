# Lab 3 - Distributed Algorithms

## Setup

From this directory, install uv and create the virtual environment:

```sh
python -m pip install uv
uv sync
```

Activate it on Windows PowerShell:

```powershell
.\.venv\Scripts\Activate.ps1
```

On macOS or Linux:

```sh
source .venv/bin/activate
```

## Usage

On Windows, mpiexec must first be installed to run MPI programs

```powershell
winget install --exact --id Microsoft.msmpi
$env:Path += ';C:\Program Files\Microsoft MPI\Bin'  # add mpiexec to your PATH for this shell session
mpiexec.exe -help  # verify installation
```

Then to run the MPI matmul program (make sure uv venv has been activated first as well).

```sh
mpiexec -n <NUM_NODES> python mpi_sort.py <N>
```

## SDSC Usage
First ssh into SDSC Expanse servers, cd into the lab directory, then run:

```powershell
module load slurm
sbatch jobscript_merge.sh
```

The slurm '.out' files are in the following format "mergesort.%j.%N.out" with %j being the job number, and %n being the node name. If you can't find the file simply use command, and type it in with the actual job number and node names:
```powershell
ls *.out 
cat mergesort.jobnumber.nodename.out
```
