from mpi4py import MPI
import sys
import numpy as np
import time

def distributed_k_way_merge(all_runs, comms):
    rank = comm.Get_rank()
    process_count = comm.Get_size()

    if rank == 0:
        assignments = [[] for _ in range(process_count)]

        for run_index, run in enumerate(all_runs):
            target_rank = run_index % process_count
            assignments[target_rank].append(run)

    else:
            assignments = None
    local_runs = comm.scatter(assignments, root = 0)

    local_result = local_k_way_merge(local_runs)
    step = 1
    while step < process_count:
        group_size = 2 * step
        if rank % group_size == 0:
            partner = rank + step
            if partner < process_count:
                #recv sorted result from partner process
                recieved = comm.recv(source = partner, tag = step,)

                #recv wins round by merging both lists
                local_result = merge(local_result, recieved, )
        elif rank % group_size == step:
            partner = rank - step

            #send proc result to lower ranked processes
            comm.send(
                local_result,
                dest=partner,
                tag=step,
            )
            return None

        step *= 2
    return local_result if rank == 0 else None

#used merge sort algorithm
def local_k_way_merge(sorted_lists):
    runs = sorted_lists
    if not runs: 
        return []
    while len(runs) > 1:
        next_round = []

        #merge adj runs
        for i in range(0, len(runs), 2):
            if i + 1 < len(runs):
                next_round.append(merge(runs[i], runs[i + 1]))
            else:
                next_round.append(runs[i])

        runs = next_round

    return runs[0]
    

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
    elif int(sys.argv[1]) < 0:
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
input_checksum = None
if rank == 0:
    N = int(sys.argv[1])
    if N == 0:
        #if given array of 0
        end_prog = True

    # create the initial random arr
    rng = np.random.default_rng(seed=42)
    arr = rng.integers(0, N, size=N)  # looks to be creating int64s

    #checksum for verification
    input_checksum = int(np.sum(arr))
    #print(arr) #commented out for cleaner print on sdsc

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

# barrier
comm.Barrier()
time_start = time.time()

# tell all ranks how many elts they have so they can pre-alloc subarrs
subcount = comm.scatter(counts, root=0)

# Scatter arr to worker nodes
subarr = np.empty((subcount), dtype=np.int64)  # match arr's int64 dtype
comm.Scatterv(
    [arr, counts, starts, MPI.INT64_T] if rank == 0 else None,
    [subarr, MPI.INT64_T],
    root=0
)

subarr.sort()

# barrier
comm.Barrier()


print(f"rank {rank} sorted: {subarr}")

# you can reference my comm.Gatherv code in lab2/matmul.py (line 71) for recombining.
# lab3 won't need the "* dim" in the args tho since arr is already a 1D array. 

#mpi reduce??
print("after rank sorted")
sorted_runs = None
final_result = None
if rank == 0:
    print("rank 0")
    sorted_runs = [subarr]
    for i in range(1, size):
        buf = comm.recv(source = MPI.ANY_SOURCE)
        sorted_runs.append(buf)
        #print("finished")
        #final_result = distributed_k_way_merge(sorted_runs, comm)
    #print(final_result)
else:
    comm.send(subarr, dest = 0)

final_result = distributed_k_way_merge(sorted_runs, comm)

if rank == 0:
    time_end = time.time()
    print(f"Completed in {time_end - time_start:.2f} seconds")

    #check all adjacent values are in nondecresing order
    is_sorted = True
    for i in range(len(final_result) - 1):
        if final_result[i] > final_result[i + 1]:
            is_sorted = False
            break

    #check output has same number of values as input
    count_ok = (len(final_result) == N)

    #check same checksum as input
    output_checksum = int(np.sum(final_result))
    sum_ok = (output_checksum==input_checksum)

    #pass/fail status
    passed = is_sorted and count_ok and sum_ok

    final_result = (final_result)
    np_result = sorted(final_result) #auto python sort function?

    if np.array_equal(final_result, np_result):
        print("This array is sorted correctly")
        # print(np_result)
        # print(final_result)

    else:
        print("this is incorrectly sorted\n")
        print(f"np: {np_result}")
        print(f"merge: {final_result}")

    print(f"Result: N={N} ranks={size} time={time_end - time_start:.4f} pass={(passed)}")
    