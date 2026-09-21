from mpi4py import MPI
import sys
import numpy as np
import time

def mergesort(arr):
    if len(arr) <= 1:
        return arr

    mid = len(arr) //2
    leftH = arr[:mid]
    rightH = arr[mid:]

    sortedLeft = mergesort(leftH)
    sortedRight = mergesort(rightH)

    return merge(sortedLeft, sortedRight)

def merge(left, right):
    result = []
    i = 0
    j = 0

    while i < len(left) and j < len(right):
        if left[i] < right[j]:
            result.append(left[i])
            i += 1
        else:
            result.append(right[j])
            j += 1

    result.extend(left[i:])
    result.extend(right[j:])

    return result

comm = MPI.COMM_WORLD
size = comm.Get_size()
rank = comm.Get_rank()
name = MPI.Get_processor_name()

# ensure correct cli args, end program otherwise
if rank == 0:  # parent node
    end_prog = False
    if len(sys.argv) != 2:
        print("Usage: mpiexec python mpi_sort.py <N>")
        end_prog = True
    elif int(sys.argv[1]) <= 0:
        print("Error: N must be a positive integer.")
        end_prog = True
else:
    end_prog = None

end_prog = comm.bcast(end_prog, root=0)  # Broadcast end_prog to all nodes
if end_prog:
    sys.exit(1)  # Exit all nodes if there was a cli arg issue

arr = None
starts = None
counts = None
if rank == 0:
    time_start = time.time()
    N = int(sys.argv[1])

    # create the initial random arr
    rng = np.random.default_rng(seed=42)
    arr = rng.integers(0, N, size=N)  # looks to be creating int64s
    print(arr)

    # calculate how the arr will be split per node
    starts = np.zeros((size), dtype=int)
    for i in range(size):
        if i == 0:
            continue
        starts[i] = i * (N // size)

    ends = np.zeros((size), dtype=int)
    for i in range(size-1):
        ends[i] = starts[i+1]
    ends[-1] = N

    counts = ends - starts

# tell all ranks how many elts they have so they can pre-alloc subarrs
subcount = comm.scatter(counts, root=0)

# Scatter arr to worker nodes
subarr = np.empty((subcount), dtype=np.int64)  # match arr's int64 dtype
comm.Scatterv(
    [arr, counts, starts, MPI.INT64_T] if rank == 0 else None,
    [subarr, MPI.INT64_T],
    root=0
)

mergesort(subarr) 

print(f"rank {rank}: {subarr}")

# you can reference my comm.Gatherv code in lab2/matmul.py (line 71) for recombining.
# lab3 won't need the "* dim" in the args tho since arr is already a 1D array. 

if rank == 0:
    final_result = np.empty()

comm.Gatherv(
    [subarr, MPI.INT64_T],
    [final_result, counts, starts, MPI.INT64_T] if rank == 0 else None, 
    root = 0
)

if rank == 0:
    time_end = time.time()
    print(f"Completed in {time_end - time_start:.2f} seconds")

    np_result = arr.sort() #auto python sort function?
    final_result = mergesort(final_result)

    if np.equal(final_result, np_result):
        print("This array is sorted correctly")
        print(np_result)
        print(final_result)

    else:
        print("this is incorrectly sorted\n")
        print(np_result)
        print(final_result)



