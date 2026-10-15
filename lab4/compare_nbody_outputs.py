"""Compare saved outputs: python compare_nbody_outputs.py nbody.log distoct.log."""

import argparse
import math
from pathlib import Path
import re
import sys

from nbody_tester import compare_bodies, parse_final_bodies, warn_negative_positions


def read_final_bodies(path):
    output = path.read_text(encoding="utf-8-sig")
    counts = re.findall(r"^\s*Bodies:\s*(\d+)\s*$", output, re.MULTILINE)
    if len(counts) != 1:
        raise ValueError(f"{path}: expected exactly one 'Bodies: <count>' header")
    if output.count("Simulation completed.") != 1:
        raise ValueError(f"{path}: expected exactly one completed simulation")
    count = int(counts[0])
    if count <= 0:
        raise ValueError(f"{path}: body count must be positive")
    try:
        return count, parse_final_bodies(output, count)
    except ValueError as exc:
        raise ValueError(f"{path}: {exc}") from exc


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("nbody_output", type=Path)
    parser.add_argument("distoct_output", type=Path)
    parser.add_argument("--tolerance", type=float, default=0.5)
    args = parser.parse_args()
    if not math.isfinite(args.tolerance) or args.tolerance < 0:
        parser.error("tolerance must be finite and nonnegative")

    try:
        reference_count, reference = read_final_bodies(args.nbody_output)
        actual_count, actual = read_final_bodies(args.distoct_output)
        if reference_count != actual_count:
            raise ValueError(f"body counts differ: nbody={reference_count}, "
                             f"distoct={actual_count}")
    except (OSError, UnicodeError, ValueError) as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 2

    warn_negative_positions(reference, "nbody")
    warn_negative_positions(actual, "distoct")
    mismatches = compare_bodies(reference, actual, args.tolerance)
    max_difference = max(
        abs(expected - observed)
        for body_id in reference
        for expected, observed in zip(reference[body_id], actual[body_id])
    )
    if mismatches:
        print(f"FAIL: {len(mismatches)} components across "
              f"{len({item[0] for item in mismatches})} bodies exceed {args.tolerance}.")
        print("Largest differences (up to 20):")
        for body_id, component, expected, observed, difference in sorted(
            mismatches, key=lambda item: item[-1], reverse=True
        )[:20]:
            print(f"Body {body_id} {component}: nbody={expected:.8g}, "
                  f"distoct={observed:.8g}, difference={difference:.8g}")
    else:
        print(f"PASS: all {reference_count} bodies match within "
              f"{args.tolerance} per component.")
    print(f"Maximum absolute component difference: {max_difference:.8g}")
    return 1 if mismatches else 0


if __name__ == "__main__":
    sys.exit(main())
