from mpi4py import MPI
import sys
import numpy as np
import time

def distributed_k_way_merge_sort(arr, comm):
    rank = comm.Get_rank()
    size = comm.Get_size()

    counts = np.empty(size, dtype=np.int64)
    displs = None
    if rank == 0:
        N = len(arr)
        for i in range(size):
            counts[i] = N // size + (i < N % size)
        displs = np.zeros(size, dtype=np.int64)
        displs[1:] = np.cumsum(counts[:-1])
    comm.Bcast(counts, root=0)

    # print(f'rank {rank} - counts: {counts}')
    local_arr = np.empty(counts[rank], dtype=np.int64)

    comm.Scatterv(
        [arr, counts, displs, MPI.INT64_T] if rank == 0 else None,
        [local_arr, MPI.INT64_T],
        root=0
    )
    local_arr.sort()
    # print(f'Rank {rank} in dist_k_merge - local_arr: {local_arr}\n')
    
    # reduce pattern
    step = 1
    while step < size:
        group_size = 2 * step
        if rank % group_size == 0:
            partner = rank + step
            if partner < size:
                #recv sorted result from partner process
                recv_count = comm.recv(source=partner, tag=step)
                received = np.empty(recv_count, dtype=np.int64)
                comm.Recv([received, MPI.INT64_T], source = partner, tag = step,)

                #recv wins round by merging both lists
                local_arr = merge(local_arr, received)
        elif rank % group_size == step:
            partner = rank - step

            #send proc result to lower ranked processes
            comm.send(len(local_arr), dest=partner, tag=step)
            comm.Send(
                [local_arr, MPI.INT64_T],
                dest=partner,
                tag=step,
            )
            return None

        step *= 2
    return local_arr if rank == 0 else None
    
def merge(left, right):
    result_len = len(left) + len(right)
    result = np.empty(result_len, dtype=np.int64)
    i = 0
    j = 0
    result_idx = 0

    while i < len(left) and j < len(right):
        if left[i] < right[j]:
            result[result_idx] = left[i]
            i += 1
        else:
            result[result_idx] = right[j]
            j += 1

        result_idx += 1

    if i < len(left):
        result[result_idx:] = left[i:]
    elif j < len(right):
        result[result_idx:] = right[j:]

    return result

def main():
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
    unsorted_runs = None
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


    # barrier
    comm.Barrier()
    time_start = MPI.Wtime()

    final_result = distributed_k_way_merge_sort(arr, comm)

    comm.Barrier()
    time_end = MPI.Wtime()
    elapsed = time_end - time_start

    max_elapsed = comm.reduce(elapsed, op=MPI.MAX, root=0)

    if rank == 0:
        print(f"Completed in {max_elapsed:.2f} seconds")

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

        print(f"Result: N={N} ranks={size} time={max_elapsed:.4f} pass={(passed)}")

if __name__ == '__main__':
    main()        