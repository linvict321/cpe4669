CC = gcc
MPICC = mpicc

all: nbody distoct

nbody: nbody.c
	$(CC) -O2 nbody.c -o nbody -lm

distoct: distoct.c
#	$(MPICC) -g -Wall -pedantic -O0 -fsanitize=address,undefined -fno-omit-frame-pointer -o distoct distoct.c -lm -pthread
	$(MPICC) -g -Wall -pedantic -O2 -o distoct distoct.c -lm -pthread

clean:
	rm -f nbody distoct

.PHONY: all clean
