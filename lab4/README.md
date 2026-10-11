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

In WSL, the tester prefers Linux binaries `nbody` and `distoct`; on Windows,
it prefers `nbody.exe` and `distoct.exe`. If Open MPI reports insufficient slots, run:

Oversubscribed runs can test correctness, but their timings do not measure scaling on dedicated cores.
