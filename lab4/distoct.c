#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <time.h>

#define G 6.67430e-11
#define SOFTENING 1e-9
#define DT 0.01

typedef struct {
    double x, y, z;      // Position
    double vx, vy, vz;   // Velocity
    double ax, ay, az;   // Acceleration
    double mass;
} Body;

/* Function prototypes */
void initialize_bodies(Body *bodies, int n);
void compute_forces(Body *bodies, int n);
/* compute_forces() provides a clean sequential baseline with O(N²) complexity. 
Students can then replace that function with an Octree/Barnes-Hut approach*/
void update_bodies(Body *bodies, int n, double dt);
void print_bodies(Body *bodies, int n);


/*
 * Initialize bodies with random positions, velocities, and masses.
 */
void initialize_bodies(Body *bodies, int n)
{
    for (int i = 0; i < n; i++) {

        bodies[i].x = ((double)rand() / RAND_MAX) * 100.0;
        bodies[i].y = ((double)rand() / RAND_MAX) * 100.0;
        bodies[i].z = ((double)rand() / RAND_MAX) * 100.0;

        bodies[i].vx = 0.0;
        bodies[i].vy = 0.0;
        bodies[i].vz = 0.0;

        bodies[i].ax = 0.0;
        bodies[i].ay = 0.0;
        bodies[i].az = 0.0;

        bodies[i].mass =
            1.0e20 + ((double)rand() / RAND_MAX) * 1.0e20;
    }
}


/*
 * Sequential O(N^2) force calculation.
 *
 * This is the baseline implementation.
 * Later, this function can be replaced with an Octree /
 * Barnes-Hut version.
 */
void compute_forces(Body *bodies, int n)
{
    /* Reset acceleration */
    for (int i = 0; i < n; i++) {
        bodies[i].ax = 0.0;
        bodies[i].ay = 0.0;
        bodies[i].az = 0.0;
    }

    /* Compute gravitational forces */
    for (int i = 0; i < n; i++) {

        for (int j = 0; j < n; j++) {

            if (i == j)
                continue;

            double dx = bodies[j].x - bodies[i].x;
            double dy = bodies[j].y - bodies[i].y;
            double dz = bodies[j].z - bodies[i].z;

            double distance_squared =
                dx * dx +
                dy * dy +
                dz * dz +
                SOFTENING;

            double distance = sqrt(distance_squared);

            double acceleration =
                G * bodies[j].mass /
                distance_squared;

            bodies[i].ax += acceleration * dx / distance;
            bodies[i].ay += acceleration * dy / distance;
            bodies[i].az += acceleration * dz / distance;
        }
    }
}


/*
 * Update velocity and position using a simple Euler integration.
 */
void update_bodies(Body *bodies, int n, double dt)
{
    for (int i = 0; i < n; i++) {

        /* Update velocity */
        bodies[i].vx += bodies[i].ax * dt;
        bodies[i].vy += bodies[i].ay * dt;
        bodies[i].vz += bodies[i].az * dt;

        /* Update position */
        bodies[i].x += bodies[i].vx * dt;
        bodies[i].y += bodies[i].vy * dt;
        bodies[i].z += bodies[i].vz * dt;
    }
}


/*
 * Print body information.
 *
 * Useful for debugging with a small number of bodies.
 */
void print_bodies(Body *bodies, int n)
{
    for (int i = 0; i < n; i++) {

        printf(
            "Body %d: "
            "pos=(%.4f, %.4f, %.4f) "
            "vel=(%.4f, %.4f, %.4f)\n",
            i,
            bodies[i].x,
            bodies[i].y,
            bodies[i].z,
            bodies[i].vx,
            bodies[i].vy,
            bodies[i].vz
        );
    }
}


int main(int argc, char *argv[])
{
    int num_bodies = 1000;
    int num_steps = 100;

    /*
     * Allow command-line arguments:
     *
     * ./nbody 10000 100
     *
     * argv[1] = number of bodies
     * argv[2] = number of simulation steps
     */
    if (argc >= 2)
        num_bodies = atoi(argv[1]);

    if (argc >= 3)
        num_steps = atoi(argv[2]);

    printf("N-Body Simulation\n");
    printf("-----------------\n");
    printf("Bodies: %d\n", num_bodies);
    printf("Steps : %d\n\n", num_steps);

    Body *bodies =
        (Body *)malloc(num_bodies * sizeof(Body));

    if (bodies == NULL) {
        fprintf(stderr, "Error allocating memory.\n");
        return EXIT_FAILURE;
    }

    srand(0);

    initialize_bodies(bodies, num_bodies);

    clock_t start = clock();

    for (int step = 0; step < num_steps; step++) {

        compute_forces(bodies, num_bodies);

        update_bodies(bodies, num_bodies, DT);

        /*
         * Uncomment for debugging.
         *
         * printf("Step %d\n", step);
         * print_bodies(bodies, num_bodies);
         */
    }

    clock_t end = clock();

    double elapsed =
        (double)(end - start) / CLOCKS_PER_SEC;

    printf("Simulation completed.\n");
    printf("Execution time: %.6f seconds\n", elapsed);

    free(bodies);

    return EXIT_SUCCESS;
}