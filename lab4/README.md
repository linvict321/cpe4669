gcc -O2 nbody.c -o nbody -lm

./nbody 1000 100

Alternatively for our lab, to compile and run use

```powershell
mpicc -g -Wall -pedantic -O0 -pthread -o distoct distoct.c -lm
mpiexec -np <NUM_RANKS> ./distoct <NUM_BODIES> <NUM_STEPS>
```

To compile with tracing for profiling:

```powershell
mpicc -o distoct distoct.c -lmpitrace
```
