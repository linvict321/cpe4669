"""Compare compiled programs: python nbody_tester.py 1000 100.

Runs distoct with 1, 2, 4, 6, and 8 ranks by default and saves execution_times.csv.

Exit codes: 0 = match, 1 = mismatch, 2 = execution or output error.
"""

import argparse
import csv
import math
from pathlib import Path
import re
import subprocess
import sys


SCRIPT_DIR = Path(__file__).resolve().parent
COMPONENTS = ("x", "y", "z", "vx", "vy", "vz")
BODY_PATTERN = re.compile(
    r"Body\s+(\d+)\s*:\s*pos=\(([^)]*)\)\s*vel=\(([^)]*)\)"
)


def parse_execution_time(output):
    marker = "Simulation completed."
    if marker not in output:
        raise ValueError("missing 'Simulation completed.' marker")
    matches = re.findall(
        r"Execution time:\s*(\S+)\s+seconds", output.split(marker, 1)[1]
    )
    if len(matches) != 1:
        raise ValueError("expected exactly one final execution time")
    seconds = float(matches[0])
    if not math.isfinite(seconds) or seconds < 0:
        raise ValueError("execution time must be finite and nonnegative")
    return seconds


def parse_final_bodies(output, expected_count):
    """Ignore initial states and require one finite final state per body ID."""
    marker = "Simulation completed."
    if marker not in output:
        raise ValueError("missing 'Simulation completed.' marker")
    bodies = {}
    for match in BODY_PATTERN.finditer(output.split(marker, 1)[1]):
        body_id = int(match.group(1))
        if body_id in bodies:
            raise ValueError(f"duplicate final body ID {body_id}")
        values = tuple(
            float(value.strip())
            for group in (match.group(2), match.group(3))
            for value in group.split(",")
        )
        if len(values) != 6 or not all(math.isfinite(value) for value in values):
            raise ValueError(f"body {body_id} must have six finite components")
        bodies[body_id] = values
    expected_ids = set(range(expected_count))
    missing = expected_ids - bodies.keys()
    unexpected = bodies.keys() - expected_ids
    if missing or unexpected:
        raise ValueError(f"invalid final body IDs: missing={sorted(missing)}, "
                         f"unexpected={sorted(unexpected)}")
    return bodies


def compare_bodies(reference, actual, tolerance):
    mismatches = []
    for body_id in sorted(reference):
        for component, expected, observed in zip(
            COMPONENTS, reference[body_id], actual[body_id]
        ):
            difference = abs(expected - observed)
            if difference > tolerance:
                mismatches.append((body_id, component, expected, observed, difference))
    return mismatches


def default_executable(name):
    for filename in (name + ".exe", name):
        path = SCRIPT_DIR / filename
        if path.is_file():
            return path
    return SCRIPT_DIR / (name + ".exe")


def run_program(command, log_path, timeout):
    print("Running:", subprocess.list2cmdline(command), flush=True)
    with log_path.open("w", encoding="utf-8") as log:
        result = subprocess.run(
            command, stdout=log, stderr=subprocess.STDOUT,
            cwd=SCRIPT_DIR, timeout=timeout, check=False,
        )
    if result.returncode != 0:
        raise RuntimeError(f"exit code {result.returncode}; see {log_path}")
    return log_path.read_text(encoding="utf-8", errors="replace")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("bodies", type=int, nargs="?", default=1000)
    parser.add_argument("steps", type=int, nargs="?", default=100)
    parser.add_argument("--ranks", type=int, nargs="+", default=[1, 2, 4, 6, 8])
    parser.add_argument("--tolerance", type=float, default=0.5)
    parser.add_argument("--nbody", type=Path, default=default_executable("nbody"))
    parser.add_argument("--distoct", type=Path, default=default_executable("distoct"))
    parser.add_argument("--mpiexec", default="mpiexec")
    parser.add_argument("--timeout", type=float, default=300,
                        help="maximum seconds per program (default: 300)")
    parser.add_argument("--output-dir", type=Path,
                        default=SCRIPT_DIR / "comparison_output")
    args = parser.parse_args()
    if args.bodies <= 0 or args.steps < 0:
        parser.error("bodies must be positive and steps must be nonnegative")
    if any(rank not in (1, 2, 4, 6, 8) for rank in args.ranks):
        parser.error("ranks must be 1, 2, 4, 6, or 8")
    if not math.isfinite(args.tolerance) or args.tolerance < 0:
        parser.error("tolerance must be finite and nonnegative")
    if not math.isfinite(args.timeout) or args.timeout <= 0:
        parser.error("timeout must be finite and positive")

    failed = False
    execution_error = False
    results = []
    timing_path = args.output_dir / "execution_times.csv"

    def new_result(program, ranks):
        row = dict(program=program, ranks=ranks, bodies=args.bodies,
                   steps=args.steps, execution_seconds="", status="ERROR")
        results.append(row)
        return row

    try:
        args.output_dir.mkdir(parents=True, exist_ok=True)
        shared_args = [str(args.bodies), str(args.steps)]
        row = new_result("nbody", 1)
        output = run_program(
            [str(args.nbody.resolve()), *shared_args],
            args.output_dir / "nbody.log", args.timeout,
        )
        row["execution_seconds"] = parse_execution_time(output)
        reference = parse_final_bodies(output, args.bodies)
        row["status"] = "REFERENCE"
        print(f"nbody execution time: {row['execution_seconds']:.8g} seconds")
        for ranks in dict.fromkeys(args.ranks):
            row = new_result("distoct", ranks)
            log_path = args.output_dir / f"distoct_{ranks}_ranks.log"
            report_path = args.output_dir / f"differences_{ranks}_ranks.txt"
            report_path.unlink(missing_ok=True)
            try:
                output = run_program(
                    [args.mpiexec, "-n", str(ranks),
                     str(args.distoct.resolve()), *shared_args],
                    log_path, args.timeout,
                )
                row["execution_seconds"] = parse_execution_time(output)
                print(f"distoct ({ranks} ranks) execution time: "
                      f"{row['execution_seconds']:.8g} seconds")
                actual = parse_final_bodies(output, args.bodies)
                mismatches = compare_bodies(reference, actual, args.tolerance)
                max_difference = max(
                    abs(a - b) for body_id in reference
                    for a, b in zip(reference[body_id], actual[body_id])
                )
                if mismatches:
                    failed = True
                    lines = ["body component nbody distoct absolute_difference"]
                    lines.extend(
                        f"{body_id} {component} {a:.8g} {b:.8g} {diff:.8g}"
                        for body_id, component, a, b, diff in mismatches
                    )
                    report_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
                    row["status"] = "FAIL"
                    print(f"FAIL ({ranks} ranks): {len(mismatches)} components "
                          f"exceed {args.tolerance}; see {report_path}")
                else:
                    row["status"] = "PASS"
                    print(f"PASS ({ranks} ranks): all {args.bodies} bodies match "
                          f"within {args.tolerance} per component")
                print(f"Maximum absolute component difference: {max_difference:.8g}")
            except (OSError, RuntimeError, ValueError, subprocess.TimeoutExpired) as exc:
                execution_error = True
                print(f"ERROR ({ranks} ranks): {exc}; output: {log_path}", file=sys.stderr)
    except (OSError, RuntimeError, ValueError, subprocess.TimeoutExpired) as exc:
        print(f"ERROR: {exc}; logs: {args.output_dir}", file=sys.stderr)
        return 2
    finally:
        try:
            with timing_path.open("w", newline="", encoding="utf-8") as timing_file:
                writer = csv.DictWriter(timing_file, fieldnames=[
                    "program", "ranks", "bodies", "steps", "execution_seconds", "status"
                ])
                writer.writeheader()
                writer.writerows(results)
            print(f"Execution times saved to {timing_path}")
        except OSError as exc:
            execution_error = True
            print(f"ERROR saving execution times: {exc}", file=sys.stderr)
    return 2 if execution_error else (1 if failed else 0)


if __name__ == "__main__":
    sys.exit(main())
