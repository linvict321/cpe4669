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
    A = np.astype(np.random.rand(dim, dim), np.float64)
    B = np.astype(np.random.rand(dim, dim), np.float64)

else:  # worker nodes
    dim = comm.bcast(None, root=0)  # Receive matrix dimension from parent node
    B = np.empty((dim, dim), dtype='float64')

starts = np.zeros((size), dtype=int)
for i in range(size):
    if i == 0:
        continue
    starts[i] = i * (dim // size)

ends = np.zeros((size), dtype=int)
for i in range(size-1):
    ends[i] = starts[i+1]
ends[-1] = dim

counts = ends - starts

local_A = np.empty((ends[rank] - starts[rank], dim), dtype=np.float64)
comm.Scatterv(
    [A, counts*dim, starts*dim, MPI.DOUBLE] if rank == 0 else None,
    [local_A, MPI.DOUBLE],
    root=0
)  # Scatter matrix A to worker nodes
comm.Bcast(B, root=0)  # Broadcast whole B matrix to all nodes

local_result = local_A @ B

if rank == 0:
    final_result = np.empty((dim, dim), dtype='float64')
else:
    final_result = None

comm.Gatherv(
    [local_result, MPI.DOUBLE],
    [final_result, counts * dim, starts * dim, MPI.DOUBLE] if rank == 0 else None,
    root=0
)

if rank == 0:
    time_end = time.time()
    print(f"Matrix multiplication of {dim}x{dim} matrices with {size} nodes completed in {time_end - time_start:.2f} seconds.")
    # verify final result is correct against numpy's matmul
    np_result = A @ B
    if np.allclose(final_result, np_result):
        print("Matrix multiplication result is correct.")
    else:
        print("Matrix multiplication result is incorrect.")
        print(f'-----Numpy result-----')
        print(np_result)
        print(f'\n-----Our result-----')
        print(final_result)
