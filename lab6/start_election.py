"""Compare compiled programs: python start_election.py <N>

Runs N processes of raft.go, each representing a node.
"""

import argparse
import subprocess
from pathlib import Path
import os

BASE_PORT = 8080

def parse_args():
    parser = argparse.ArgumentParser(description="Script for starting up N nodes for Raft election protocol.")
    parser.add_argument("N", help="Number of nodes to start", type=int)

    args = parser.parse_args()
    if args.N < 1:
        parser.error("N must be at least 1")

    return args

def main():
    args = parse_args()

    # compile the prog once
    executable = Path(
        "raft.exe" if os.name == "nt" else "raft"
    ).resolve()
    subprocess.run(
        ["go", "build", "-o", str(executable), "raft.go"],
        check=True,
    )

    processes = []
    try:
        for i in range(args.N):
            processes.append(subprocess.Popen([
                str(executable), str(i), str(args.N)
            ]))
            print(f'Started process with id {i} at port {BASE_PORT+i}, with pid {processes[i].pid}')

        for process in processes:
            process.wait()

    except KeyboardInterrupt:
        pass
    finally:
        for process in processes:
            if process.poll() is None:  # poll returns exit code if finished, otherwise None
                process.terminate()
        for process in processes:
            process.wait()


if __name__ == '__main__':
    main()