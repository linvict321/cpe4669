# Lab 2 - Distributed Algorithms

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
Replace <NUM_NODES> and <MATRIX_DIM> appropriately.

```sh
mpiexec -n <NUM_NODES> python matmul.py <MATRIX_DIM>
```

Note both A and B will both be random matrices of floats with shape (MATRIX_DIM, MATRIX_DIM)
