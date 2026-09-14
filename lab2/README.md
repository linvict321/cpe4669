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

```sh
mpiexec -n <NUM_NODES> python matmul.py <MATRIX_DIM>
```

Note both A and B will both be random matrices of float64's with shape (MATRIX_DIM, MATRIX_DIM)


# RPC Setup

```sh
cd lab2
cd rpc_parallel

```
open up 2 terminals (or split terminal)

for client terminal:
```sh
cd client
go run client.go
```

for server terminal:
```sh
cd server
go run server.go
```

to change num of worker nodes: go into client.go and change numClients

the searched word is in client.go 's Target; the file being searched is RPC_text.txt in the server folder