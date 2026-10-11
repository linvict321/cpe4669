gcc -O2 nbody.c -o nbody -lm

./nbody 1000 100

Compare the compiled sequential and distributed programs with the same inputs:
Distributed implementation is run with {1, 2, 4, 6, 8} ranks

```powershell
python nbody_tester.py <NUM_BODIES> <NUM_STEPS>
```

Alternatively for our lab, to compile and run use

```powershell
mpicc -g -Wall -pedantic -O0 -pthread -o distoct.exe distoct.c -lm
mpiexec -np <NUM_RANKS> ./distoct.exe <NUM_BODIES> <NUM_STEPS>
```

To compile with tracing for profiling:

```powershell
mpicc -o distoct distoct.c -lmpitrace
```
