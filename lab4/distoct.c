#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <time.h>
#include <mpi.h>
#include <string.h>

#define G 6.67430e-11
#define SOFTENING 1e-1  // was originally 1e-9 but this was negligible and causing numerical "slingshotting" for bodies close together
#define DT 0.01
#define INITIAL_WIDTH 100.0
#define MASS_ORDER 1.0e10
#define ACCURACY 0.5  // google said this was the "sweet spot", though [0.5, 1] is a normal range

const int MAX_STRING = 100;

// currently only intended for up to 8 ranks at once due to the octree

typedef struct {
    double x, y, z;      // Position
    double vx, vy, vz;   // Velocity
    double ax, ay, az;   // Acceleration
    double mass;
} Body;

typedef struct {
    // the actual space min/max
    double minx, miny, minz;
    double maxx, maxy, maxz;

    // the limits of what the box can spatially expand to
    double lminx, lminy, lminz;
    double lmaxx, lmaxy, lmaxz;
} Box;

typedef struct {
    double mass; // sum of bodies' masses
    double COMx, COMy, COMz; // weighted COM position
    Box *bbox; // node bounding box
    double sx, sy, sz; // spatial center position
    double width; // cell spatial width
    Body *bodies; // bodies contained within this node
    struct OctreeNode *childNodes; // NULL if this OctreeNode is a leaf node
} OctreeNode;

/* Function prototypes */
void initialize_bodies(Body *bodies, int n);
void compute_forces(Body *bodies, int n);
void compute_forces_dist(Body *bodies, int n);
/* compute_forces() provides a clean sequential baseline with O(N²) complexity. 
Students can then replace that function with an Octree/Barnes-Hut approach*/
void update_bodies(Body *bodies, int n, double dt);
void print_bodies(Body *bodies, int n);
int acceptance_test(double cell_size, double distance, double accuracy);
Box * divide_3d_space(double width);
void get_idxs_for_each_rank(int *starts, int *ends, int n, int size);
void handle_mpi_errcode(int errcode);


/*
 * Initialize bodies with random positions, velocities, and masses.
 */
void initialize_bodies(Body *bodies, int n)
{
    for (int i = 0; i < n; i++) {

        // x,y,z are all doubles within [0, 100]
        bodies[i].x = ((double)rand() / RAND_MAX) * INITIAL_WIDTH;
        bodies[i].y = ((double)rand() / RAND_MAX) * INITIAL_WIDTH;
        bodies[i].z = ((double)rand() / RAND_MAX) * INITIAL_WIDTH;

        // all bodies start at rest with no acceleration
        bodies[i].vx = 0.0;
        bodies[i].vy = 0.0;
        bodies[i].vz = 0.0;

        bodies[i].ax = 0.0;
        bodies[i].ay = 0.0;
        bodies[i].az = 0.0;

        // mass bounded between [MASS_ORDER, 2 * MASS_ORDER]
        bodies[i].mass =
            MASS_ORDER + ((double)rand() / RAND_MAX) * MASS_ORDER;
    }
}

 // writing new version right here
void compute_forces_dist(Body *bodies, int n){
    /* Reset acceleration */
    for (int i = 0; i < n; i++) {
        bodies[i].ax = 0.0;
        bodies[i].ay = 0.0;
        bodies[i].az = 0.0;
    }
}

/* Sequential O(N^2) force calculation. */
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

            // the acceleration of body i due to j
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

void get_idxs_for_each_rank(int *starts, int *ends, int n, int size) {
    for (int r = 0; r < size; r++) {
        starts[r] = (r * n) / size;
        if (r == size - 1) {
            int last_end = ((r+1) * n) / size;
            ends[r] = n > last_end ? n : last_end;
        } else {
            ends[r] = ((r+1) * n) / size;
        }
    }
}

double calc_OctreeNode_extrema(Body* bodies, int num_bodies) {
    double max;
}

void build_local_octree(OctreeNode* node, double width, Body* bodies, int num_bodies) {

}

void build_octree(OctreeNode* node, double width, Body* bodies, int num_bodies) {


}

Box* divide_3d_space(double width) {
    Box *boxes = (Box*)malloc(sizeof(Box) * 8); // 8 octants
    int i = 0;
    boxes[i].minx = 0; boxes[i].miny = 0; boxes[i].minz = 0;
    boxes[i].maxx = width / 2; boxes[i].maxy = width / 2; boxes[i].maxz = width / 2;
    boxes[i].lminx = -INFINITY; boxes[i].lminy = -INFINITY; boxes[i].lminz = -INFINITY;
    boxes[i].lmaxx = width / 2; boxes[i].lmaxy = width / 2; boxes[i].lmaxz = width / 2;
    i++; // now i = i
    boxes[i].minx = width / 2; boxes[i].miny = 0; boxes[i].minz = 0;
    boxes[i].maxx = width; boxes[i].maxy = width / 2; boxes[i].maxz = width / 2;
    boxes[i].lminx = width / 2; boxes[i].lminy = -INFINITY; boxes[i].lminz = -INFINITY;
    boxes[i].lmaxx = INFINITY; boxes[i].lmaxy = width / 2; boxes[i].lmaxz = width / 2;
    i++; // now i = 2
    boxes[i].minx = width / 2; boxes[i].miny = width / 2; boxes[i].minz = 0;
    boxes[i].maxx = width; boxes[i].maxy = width; boxes[i].maxz = width / 2;
    boxes[i].lminx = width / 2; boxes[i].lminy = width / 2; boxes[i].lminz = -INFINITY;
    boxes[i].lmaxx = INFINITY; boxes[i].lmaxy = INFINITY; boxes[i].lmaxz = width / 2;
    i++; // now i = 3
    boxes[i].minx = 0; boxes[i].miny = width / 2; boxes[i].minz = 0;
    boxes[i].maxx = width / 2; boxes[i].maxy = width; boxes[i].maxz = width / 2;
    boxes[i].lminx = -INFINITY; boxes[i].lminy = width / 2; boxes[i].lminz = -INFINITY;
    boxes[i].lmaxx = width / 2; boxes[i].lmaxy = INFINITY; boxes[i].lmaxz = width / 2;
    
    i++; // now i = 4
    boxes[i].minx = 0; boxes[i].miny = width / 2; boxes[i].minz = width / 2;
    boxes[i].maxx = width / 2; boxes[i].maxy = width; boxes[i].maxz = width;
    boxes[i].lminx = -INFINITY; boxes[i].lminy = width / 2; boxes[i].lminz = width / 2;
    boxes[i].lmaxx = width / 2; boxes[i].lmaxy = INFINITY; boxes[i].lmaxz = INFINITY;
    i++; // now i = 5
    boxes[i].minx = width / 2; boxes[i].miny = width / 2; boxes[i].minz = width / 2;
    boxes[i].maxx = width; boxes[i].maxy = width; boxes[i].maxz = width;
    boxes[i].lminx = width / 2; boxes[i].lminy = width / 2; boxes[i].lminz = width / 2;
    boxes[i].lmaxx = INFINITY; boxes[i].lmaxy = INFINITY; boxes[i].lmaxz = INFINITY;
    i++; // now i = 6
    boxes[i].minx = width / 2; boxes[i].miny = 0; boxes[i].minz = width / 2;
    boxes[i].maxx = width; boxes[i].maxy = width / 2; boxes[i].maxz = width;
    boxes[i].lminx = width / 2; boxes[i].lminy = -INFINITY; boxes[i].lminz = width / 2;
    boxes[i].lmaxx = INFINITY; boxes[i].lmaxy = width / 2; boxes[i].lmaxz = INFINITY;
    i++; // now i = 7
    boxes[i].minx = 0; boxes[i].miny = 0; boxes[i].minz = width / 2;
    boxes[i].maxx = width / 2; boxes[i].maxy = width / 2; boxes[i].maxz = width;
    boxes[i].lminx = -INFINITY; boxes[i].lminy = -INFINITY; boxes[i].lminz = width / 2;
    boxes[i].lmaxx = width / 2; boxes[i].lmaxy = width / 2; boxes[i].lmaxz = INFINITY;
    
    return boxes;
}

void print_boxes(Box *boxes, int n)
{
    for (int i = 0; i < n; i++) {

        printf(
            "Octant %d: "
            "min=(%.4f, %.4f, %.4f) "
            "max=(%.4f, %.4f, %.4f) "
            "lmin=(%.4f, %.4f, %.4f) "
            "lmax=(%.4f, %.4f, %.4f)\n",
            i,
            boxes[i].minx,
            boxes[i].miny,
            boxes[i].minz,
            boxes[i].maxx,
            boxes[i].maxy,
            boxes[i].maxz,
            boxes[i].lminx,
            boxes[i].lminy,
            boxes[i].lminz,
            boxes[i].lmaxx,
            boxes[i].lmaxy,
            boxes[i].lmaxz
        );
    }
}

/*
* returns 1 if far away -> use one aggregate force
* returns 0 if too close -> open the octree cell and inspect children
*/
int acceptance_test(double cell_size, double distance, double accuracy) {
    return (cell_size / distance) < accuracy;
}

/* 
* MPI error handling snippet from
* https://www.paulnorvig.com/guides/introduction-to-mpi-with-c.html
*/
void handle_mpi_errcode(int errcode) {
    if (errcode != MPI_SUCCESS) {
        char err_string[MPI_MAX_ERROR_STRING];
        int resultlen;
        MPI_Error_string(errcode, err_string, &resultlen);
        fprintf(stderr, err_string);
        MPI_Finalize();
        exit(1);
    }
}

int main(int argc, char *argv[])
{
    int errcode;
    errcode = MPI_Init(&argc, &argv);
    handle_mpi_errcode(errcode);

    // char greeting[MAX_STRING];
    int size; //mum of processes
    int rank;
    MPI_Comm_size(MPI_COMM_WORLD, &size);
    MPI_Comm_rank(MPI_COMM_WORLD, &rank);

    if (size > 8) {
        if (rank == 0) {
            fprintf(stderr, "Cannot distribute across more than 8 ranks. Aborting all ranks.\n");
        }
        MPI_Finalize();
        exit(1);
    }

    // default values
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

    if (rank == 0) {
        printf("N-Body Simulation\n");
        printf("-----------------\n");
        printf("Bodies: %d\n", num_bodies);
        printf("Steps : %d\n\n", num_steps);
    }

    
    srand(0);
    clock_t start, end;
    Body *bodies;
    if(rank == 0){
        bodies = (Body *)malloc(num_bodies * sizeof(Body));
        
        if (bodies == NULL) {
            fprintf(stderr, "Error allocating memory.\n");
            MPI_Finalize();
            return EXIT_FAILURE;
        }
        
        
        initialize_bodies(bodies, num_bodies);
        
        start = clock();
        
        
        
        
    }
    
    Box *boxes = divide_3d_space(INITIAL_WIDTH);
    
    // calc how body ownership will be distributed across ranks
    // each rank is responsible for bodies [starts[r], ends[r])
    int starts[size]; // inclusive
    int ends[size]; // exclusive
    get_idxs_for_each_rank(starts, ends, 8, size); // 8 octants
    printf("Rank %d has octants [%d, %d)\n", rank, starts[rank], ends[rank]);
    
    if (rank == 0) {
        print_boxes(boxes, 8);
    }

    int num_local_bodies = 25;  // adjust so that this becomes however many bodies are in the rank's area
    Body *local_bodies = (Body *)malloc(sizeof(Body) * num_local_bodies);


    // for (int step = 0; step < num_steps; step++) {
    //     // recalculate local octree

    //     compute_forces(bodies, num_bodies);
    // errcode = MPI_Barrier(MPI_COMM_WORLD);
    // handle_mpi_errcode(errcode);

    //     update_bodies(bodies, num_bodies, DT);

    //     /*
    //      * Uncomment for debugging.
    //      *
    //      * printf("Step %d\n", step);
    //      * print_bodies(bodies, num_bodies);
    //      */
    // }

    if (rank == 0) {
        end = clock();

        double elapsed =
            (double)(end - start) / CLOCKS_PER_SEC;

        printf("Simulation completed.\n");
        printf("Execution time: %.6f seconds\n", elapsed);

        /* check against sequential by writing the print_bodies
         * output into a file, then diff'ing that file against
         * the the octree print_bodies output written to a different file.
         */ 
        print_bodies(bodies, num_bodies);

        free(bodies);
    }
    free(boxes);
    free(local_bodies);

    MPI_Finalize();

    return EXIT_SUCCESS;
}