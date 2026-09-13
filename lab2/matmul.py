from mpi4py import MPI
import sys
import numpy as np
import time

comm = MPI.COMM_WORLD
size = comm.Get_size()
rank = comm.Get_rank()
name = MPI.Get_processor_name()

A = None
if rank == 0:  # parent node
    end_prog = False
    if len(sys.argv) != 2:
        print("Usage: python matmul.py <MATRIX_DIM>")
        end_prog = True
    elif int(sys.argv[1]) <= 0:
        print("Error: MATRIX_DIM must be a positive integer.")
        end_prog = True
    elif int(sys.argv[1]) % size != 0:
        print(f"Error: MATRIX_DIM must be divisible by the number of nodes ({size}).")
        end_prog = True
else:
    end_prog = None

end_prog = comm.bcast(end_prog, root=0)  # Broadcast end_prog to all nodes
if end_prog:
    sys.exit(1)  # Exit all nodes if there was a cli arg issue

final_result = None
# now that errors are handled, do the actual matrix inits
if rank == 0:
    time_start = time.time()
    # Read matrix dimensions from command line arguments
    dim = int(sys.argv[1])
    comm.bcast(dim, root=0)  # Broadcast matrix dimension to all nodes

    # Generate random matrices A and B
    A = np.random.rand(dim, dim)
    B = np.random.rand(dim, dim)

else:  # worker nodes
    dim = comm.bcast(None, root=0)  # Receive matrix dimension from parent node
    B = np.empty((dim, dim), dtype='float64')

local_row_start = rank * (dim // size)  # Starting row index for this worker node
local_row_end = (rank + 1) * (dim // size)  # Ending row index for this worker node
local_A = np.empty((local_row_end - local_row_start, dim), dtype='float64')
comm.Scatter(A, local_A, root=0)  # Scatter matrix A to worker nodes
comm.Bcast(B, root=0)  # Broadcast whole B matrix to all nodes

# sys.stdout.write(f"Node {rank} on {name} received matrix A with shape {local_A.shape} and matrix B with shape {B.shape}\n")
local_result = local_A @ B

if rank == 0:
    final_result = np.empty((dim, dim), dtype='float64')
comm.Gather(local_result, final_result, root=0)
if rank == 0:
    time_end = time.time()
    print(f"Matrix multiplication of {dim}x{dim} matrices with {size} nodes completed in {time_end - time_start:.2f} seconds.")
    # verify final result is correct against numpy's matmul
    np_result = A @ B
    if np.allclose(final_result, np_result):
        print("Matrix multiplication result is correct.")
    else:
        print("Matrix multiplication result is incorrect.")