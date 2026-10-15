# Lab 4 - Distributed Barnes-Hut N-Body Simulation

To compile both the given sequential and distributed implementations, run:

```bash
make
```

To run just the sequential nbody sim

```bash
./nbody <NUM_BODIES> <NUM_STEPS>
```

To run just the distributed sim

```bash
mpiexec -n <NUM_RANKS> ./distoct <NUM_BODIES> <NUM_STEPS>
```

Note that distoct only works with NUM_RANKS $\in$ {1, 2, 4, 6, 8}

To compare the compiled sequential and distributed programs with the same inputs:
Distributed implementation is run with {1, 2, 4, 6, 8} ranks, run:

```bash
python nbody_tester.py <NUM_BODIES> <NUM_STEPS>
```

If getting errors when running with the higher number of ranks, run instead using:

```bash
python nbody_tester.py <NUM_BODIES> <NUM_STEPS> --oversubscribe
```

Can also add the `--nbody_use_hard_boundaries` flag for the implementations to still match when a body was trying to go out of the original space.

To compare two saved program outputs without running either simulation:

```bash
python compare_nbody_outputs.py nbody_output.txt distoct_output.txt
```

In WSL, the tester prefers Linux binaries `nbody` and `distoct`; on Windows,
it prefers `nbody.exe` and `distoct.exe`. If Open MPI reports insufficient slots, run:

Oversubscribed runs can test correctness, but their timings do not measure scaling on dedicated cores.
