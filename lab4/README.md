gcc -O2 nbody.c -o nbody -lm

./nbody 1000 100

Alternatively for our lab, to compile and run use

```powershell
mpicc -O2 -o distoct distoct.c
mpiexec -np <NUM_RANKS> ./distoct
```

To compile with tracing for profiling:

```powershell
mpicc -o distoct distoct.c -lmpitrace
```
