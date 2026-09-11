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

```sh
mpiexec -n 4 python matmul.py
```
