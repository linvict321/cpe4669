#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <time.h>
#include <mpi.h>
#include <string.h>
#include <pthread.h>

#define G 6.67430e-11
#define SOFTENING 1e-1  // was originally 1e-9 but this was negligible and causing numerical "slingshotting" for bodies close together
#define DT 0.01
#define INITIAL_WIDTH 100.0
#define MASS_ORDER 1.0e10
#define ACCURACY 0.5  // google said this was the "sweet spot", though [0.5, 1] is a normal range
// #define MAX_DEPTH 100

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
    // no longer needed since not letting bodies outside initial global space
    double lminx, lminy, lminz;
    double lmaxx, lmaxy, lmaxz;
} Box;

typedef struct OctreeNode {
    double mass; // sum of bodies' masses
    double COMx, COMy, COMz; // weighted COM position
    double max_dim;
    Box bbox; // node bounding box
    struct OctreeNode **childNodes; // NULL if this OctreeNode is a leaf node
} OctreeNode;

/* Function prototypes */
void initialize_bodies(Body *bodies, int n);
void compute_forces(Body *bodies, int n);
void update_bodies(Body *bodies, int n, double dt, int width);
void print_bodies(Body *bodies, int n);

void compute_forces_dist(Body *bodies, int n, OctreeNode *local_octree, OctreeNode *octrees, int rank, int size);
double compute_distance(double dx, double dy, double sz);
int acceptance_test(double cell_size, double distance, double accuracy);
void get_idxs_for_each_rank(int *starts, int *ends, int n, int size);
void handle_mpi_errcode(int errcode);
int body_is_inside_box(Box box, Body* body);
OctreeNode* build_local_octree(Box box, Body* bodies, int num_bodies);
void printOctree(OctreeNode *node, int depth);
Box init_global_space(double width);
Box* divide_box_into_octants(Box parent);
Box merge_boxes(Box *boxes, int num_boxes);
double get_box_max_dim(Box box);
void send_body(Body body, int dst_rank);


/*
 * Initialize bodies with random positions, velocities, and masses.
 */
void initialize_bodies(Body *bodies, int n) {
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

double get_box_max_dim(Box box) {
    double wx = box.maxx - box.minx;
    double wy = box.maxy - box.miny;
    double wz = box.maxz - box.minz;

    double max_dim = wx;
    if (wy > max_dim) max_dim = wy;
    if (wz > max_dim) max_dim = wz;

    return max_dim;
}

 // writing new version right here
void compute_forces_dist(Body *bodies, int n, OctreeNode *local_octree, OctreeNode *octrees, int rank, int size) {
    /* Reset acceleration */
    for (int i = 0; i < n; i++) {
        bodies[i].ax = 0.0;
        bodies[i].ay = 0.0;
        bodies[i].az = 0.0;
    }
    
    // TODO: WRITE ALG FOR ANY OCTREE NODE to a body
    // maybe need to have separate one for same vs diff rank octree node?
    OctreeNode *node;
    double dx, dy, dz, distance, distance_squared, acceleration;
    for (int i = 0; i < n; i++) {
        // for each body, go through all octrees (resursively as needed)
        // to calc the forces on it --> its accels
        for (int r = 0; r < size; r++) {
            if (r == rank) {
                node = local_octree;
            } else {
                node = octrees + r;
            }

            dx = node->COMx - bodies[i].x,
            dy = node->COMy - bodies[i].y,
            dz = node->COMz - bodies[i].z,
            distance_squared = dx * dx + dy * dy + dz * dz + SOFTENING;
            distance = sqrt(distance_squared);

            if (acceptance_test(node->max_dim, distance, ACCURACY) == 1) {
                // the acceleration of body i due to node (treated as one body)
                double acceleration = G * node->mass / distance_squared;

                bodies[i].ax += acceleration * dx / distance;
                bodies[i].ay += acceleration * dy / distance;
                bodies[i].az += acceleration * dz / distance;
            } else {
                // do i need to adjust this to be recursive?
                // or i could put into a while loop with the 
                // acceptance test as the condition

            }
        }
    }
}

/* Sequential O(N^2) force calculation. */
void compute_forces(Body *bodies, int n) {
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
void update_bodies(Body *bodies, int n, double dt, int width) {
    for (int i = 0; i < n; i++) {

        /* Update velocity */
        bodies[i].vx += bodies[i].ax * dt;
        bodies[i].vy += bodies[i].ay * dt;
        bodies[i].vz += bodies[i].az * dt;

        /* Update position */
        bodies[i].x += bodies[i].vx * dt;
        bodies[i].y += bodies[i].vy * dt;
        bodies[i].z += bodies[i].vz * dt;
        
        // if bodies are outside the global bounds, reset to the "wall"
        if (bodies[i].x < 0) bodies[i].x = 0;
        if (bodies[i].y < 0) bodies[i].y = 0;
        if (bodies[i].z < 0) bodies[i].z = 0;

        if (bodies[i].x > width) bodies[i].x = width;
        if (bodies[i].y > width) bodies[i].y = width;
        if (bodies[i].z > width) bodies[i].z = width;
    }
}

/*
 * Print body information.
 *
 * Useful for debugging with a small number of bodies.
 */
void print_bodies(Body *bodies, int n) {
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

void printOctree(OctreeNode *node, int depth) {
    if (node == NULL) {
        return;
    }

    // indentation
    for (int i = 0; i < depth; i++) {
        printf("  ");
    }

    printf("Node:\n");

    for (int i = 0; i < depth; i++) {
        printf("  ");
    }
    printf("  mass = %.6f\n", node->mass);

    for (int i = 0; i < depth; i++) {
        printf("  ");
    }
    printf("  COM = (%.6f, %.6f, %.6f)\n",
           node->COMx, node->COMy, node->COMz);

    for (int i = 0; i < depth; i++) {
        printf("  ");
    }
    printf("  bbox = "
            "x[%.3f, %.3f] "
            "y[%.3f, %.3f] "
            "z[%.3f, %.3f]\n",
            node->bbox.minx, node->bbox.maxx,
            node->bbox.miny, node->bbox.maxy,
            node->bbox.minz, node->bbox.maxz);

    for (int i = 0; i < depth; i++) {
        printf("  ");
    }
    printf("  limits = "
            "x[%.3f, %.3f] "
            "y[%.3f, %.3f] "
            "z[%.3f, %.3f]\n",
            node->bbox.lminx, node->bbox.lmaxx,
            node->bbox.lminy, node->bbox.lmaxy,
            node->bbox.lminz, node->bbox.lmaxz);

    // leaf node
    if (node->childNodes == NULL) {
        for (int i = 0; i < depth; i++) {
            printf("  ");
        }
        printf("  LEAF\n");
        return;
    }

    // recursively print all 8 children
    for (int i = 0; i < 8; i++) {
        if (node->childNodes[i] != NULL) {
            for (int j = 0; j < depth; j++) {
                printf("  ");
            }

            printf("  Child %d:\n", i);

            printOctree(node->childNodes[i], depth + 1);
        }
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

void calc_OctreeNode_extrema(Box box, Body* bodies, int num_bodies) {
    if (bodies == NULL || num_bodies < 1) {
        return;
    }
    
    double x, y, z;
    box.minx = box.maxx = bodies[0].x;
    box.miny = box.maxy = bodies[0].y;
    box.minz = box.maxz = bodies[0].z;
    for (int i = 1; i < num_bodies; i++) {
        x = bodies[i].x; y = bodies[i].y; z = bodies[i].z;
        if (x < box.minx) box.minx = x;
        if (x > box.maxx) box.maxx = x;
        if (y < box.miny) box.miny = y;
        if (y > box.maxy) box.maxy = y;
        if (z < box.minz) box.minz = z;
        if (z > box.maxz) box.maxz = z;
    }
}

void calc_global_spatial_extrema(Box global, Box* locals, int num_locals) {
    if (num_locals < 1 || locals == NULL) {
        return;
    }
    global.minx = locals[0].minx; global.maxx = locals[0].maxx;
    global.miny = locals[0].miny; global.maxy = locals[0].maxy;
    global.minz = locals[0].minz; global.maxz = locals[0].maxz;
    for (int i = 1; i < num_locals; i++) {
        if (locals[i].minx < global.minx) global.minx = locals[i].minx;
        if (locals[i].maxx > global.maxx) global.maxx = locals[i].maxx;
        if (locals[i].miny < global.miny) global.miny = locals[i].miny;
        if (locals[i].maxy > global.maxy) global.maxy = locals[i].maxy;
        if (locals[i].minz < global.minz) global.minz = locals[i].minz;
        if (locals[i].maxz > global.maxz) global.maxz = locals[i].maxz;
    }
}

/*
* Returns 1 if body is within box, else returns 0
*/
int body_is_inside_box(Box box, Body* body) {
    if (body->x < box.minx || box.maxx < body->x) return 0;
    if (body->y < box.miny || box.maxy < body->y) return 0;
    if (body->z < box.minz || box.maxz < body->z) return 0;
    return 1;
}

/*
* TODO: adjust so only bodies that are within the box
* should be passed in for bodies when recursing to the next depth
*/
OctreeNode* build_local_octree(Box box, Body* bodies, int num_bodies) {
    OctreeNode *node = (OctreeNode *)calloc(1, sizeof(OctreeNode));
    node->bbox = box;
    node->max_dim = get_box_max_dim(box);

    // calc total mass and COM
    int bodies_in_node = 0;
    for (int i = 0; i < num_bodies; i++) {
        if (body_is_inside_box(box, bodies + i) == 1) {
            bodies_in_node++;
            node->mass += bodies[i].mass;
            node->COMx += bodies[i].mass * bodies[i].x;
            node->COMy += bodies[i].mass * bodies[i].y;
            node->COMz += bodies[i].mass * bodies[i].z;
        }
    }
    if (node->mass > 0) {
        node->COMx /= node->mass;
        node->COMy /= node->mass;
        node->COMz /= node->mass;
    }

    // break up node into another 8 octree nodes if more than 1 body is within
    if (bodies_in_node > 1) {
        Box *child_boxes = divide_box_into_octants(box);
        node->childNodes = (OctreeNode **)calloc(8, sizeof(OctreeNode *));
        for (int i = 0; i < 8; i++) {
            node->childNodes[i] = build_local_octree(child_boxes[i], bodies, num_bodies);
        }
        free(child_boxes);
    } else {
        node->childNodes = NULL;
    }

    return node;
}

void free_octree(OctreeNode *node) {
    if (node == NULL) return;
    if (node->childNodes == NULL) { // leaf node, free
        free(node);
    } else {
        for (int i = 0; i < 8; i++) {
            free_octree(node->childNodes[i]);
        }
    }
}

Box* divide_box_into_octants(Box parent) {
    Box *children = (Box*)calloc(8, sizeof(Box)); // 8 octants
    double cx = (parent.minx + parent.maxx) / 2; 
    double cy = (parent.miny + parent.maxy) / 2;
    double cz = (parent.minz + parent.maxz) / 2;
    
    int i = 0;
    children[i].minx = parent.minx; children[i].miny = parent.miny; children[i].minz = parent.minz;
    children[i].maxx = cx; children[i].maxy = cy; children[i].maxz = cz;
    children[i].lminx = parent.lminx; children[i].lminy = parent.lminy; children[i].lminz = parent.lminz;
    children[i].lmaxx = cx; children[i].lmaxy = cy; children[i].lmaxz = cz;
    i++; // now i = i
    children[i].minx = cx; children[i].miny = parent.miny; children[i].minz = parent.minz;
    children[i].maxx = parent.maxx; children[i].maxy = cy; children[i].maxz = cz;
    children[i].lminx = cx; children[i].lminy = parent.lminy; children[i].lminz = parent.lminz;
    children[i].lmaxx = parent.lmaxx; children[i].lmaxy = cy; children[i].lmaxz = cz;
    i++; // now i = 2
    children[i].minx = cx; children[i].miny = cy; children[i].minz = parent.minz;
    children[i].maxx = parent.maxx; children[i].maxy = parent.maxy; children[i].maxz = cz;
    children[i].lminx = cx; children[i].lminy = cy; children[i].lminz = parent.lminz;
    children[i].lmaxx = parent.lmaxx; children[i].lmaxy = parent.lmaxy; children[i].lmaxz = cz;
    i++; // now i = 3
    children[i].minx = parent.minx; children[i].miny = cy; children[i].minz = parent.minz;
    children[i].maxx = cx; children[i].maxy = parent.maxy; children[i].maxz = cz;
    children[i].lminx = parent.lminx; children[i].lminy = cy; children[i].lminz = parent.lminz;
    children[i].lmaxx = cx; children[i].lmaxy = parent.lmaxy; children[i].lmaxz = cz;

    i++; // now i = 4
    children[i].minx = parent.minx; children[i].miny = cy; children[i].minz = cz;
    children[i].maxx = cx; children[i].maxy = parent.maxy; children[i].maxz = parent.maxz;
    children[i].lminx = parent.lminx; children[i].lminy = cy; children[i].lminz = cz;
    children[i].lmaxx = cx; children[i].lmaxy = parent.lmaxy; children[i].lmaxz = parent.lmaxz;
    i++; // now i = 5
    children[i].minx = cx; children[i].miny = cy; children[i].minz = cz;
    children[i].maxx = parent.maxx; children[i].maxy = parent.maxy; children[i].maxz = parent.maxz;
    children[i].lminx = cx; children[i].lminy = cy; children[i].lminz = cz;
    children[i].lmaxx = parent.lmaxx; children[i].lmaxy = parent.lmaxy; children[i].lmaxz = parent.lmaxz;
    i++; // now i = 6
    children[i].minx = cx; children[i].miny = parent.miny; children[i].minz = cz;
    children[i].maxx = parent.maxx; children[i].maxy = cy; children[i].maxz = parent.maxz;
    children[i].lminx = cx; children[i].lminy = parent.lminy; children[i].lminz = cz;
    children[i].lmaxx = parent.lmaxx; children[i].lmaxy = cy; children[i].lmaxz = parent.lmaxz;
    i++; // now i = 7
    children[i].minx = parent.minx; children[i].miny = parent.miny; children[i].minz = cz;
    children[i].maxx = cx; children[i].maxy = cy; children[i].maxz = parent.maxz;
    children[i].lminx = parent.lminx; children[i].lminy = parent.lminy; children[i].lminz = cz;
    children[i].lmaxx = cx; children[i].lmaxy = cy; children[i].lmaxz = parent.lmaxz;
    
    return children;
}

/*
* Tested -- Only works as intended if the boxes being merged form a rectangular prism
*/
Box merge_boxes(Box *boxes, int num_boxes) {
    if (boxes == NULL || num_boxes < 1) {
        fprintf(stderr, "Error, merge_boxes received NULL/zero arguments\n");
        exit(EXIT_FAILURE);
    }
    
    Box merged;
    merged.minx = boxes[0].minx; merged.maxx = boxes[0].maxx;
    merged.miny = boxes[0].miny; merged.maxy = boxes[0].maxy;
    merged.minz = boxes[0].minz; merged.maxz = boxes[0].maxz;
    merged.lminx = boxes[0].lminx; merged.lmaxx = boxes[0].lmaxx;
    merged.lminy = boxes[0].lminy; merged.lmaxy = boxes[0].lmaxy;
    merged.lminz = boxes[0].lminz; merged.lmaxz = boxes[0].lmaxz;

    for (int i = 1; i < num_boxes; i++) {
        if (boxes[i].minx < merged.minx) merged.minx = boxes[i].minx;
        if (boxes[i].maxx > merged.maxx) merged.maxx = boxes[i].maxx;
        if (boxes[i].miny < merged.miny) merged.miny = boxes[i].miny;
        if (boxes[i].maxy > merged.maxy) merged.maxy = boxes[i].maxy;
        if (boxes[i].minz < merged.minz) merged.minz = boxes[i].minz;
        if (boxes[i].maxz > merged.maxz) merged.maxz = boxes[i].maxz;

        if (boxes[i].lminx < merged.lminx) merged.lminx = boxes[i].lminx;
        if (boxes[i].lmaxx > merged.lmaxx) merged.lmaxx = boxes[i].lmaxx;
        if (boxes[i].lminy < merged.lminy) merged.lminy = boxes[i].lminy;
        if (boxes[i].lmaxy > merged.lmaxy) merged.lmaxy = boxes[i].lmaxy;
        if (boxes[i].lminz < merged.lminz) merged.lminz = boxes[i].lminz;
        if (boxes[i].lmaxz > merged.lmaxz) merged.lmaxz = boxes[i].lmaxz;
    }
    return merged;
}

Box init_global_space(double width) {
    Box global;
    global.minx = global.miny = global.minz = 0;
    global.maxx = global.maxy = global.maxz = width;
    global.lminx = global.lminy = global.lminz = -INFINITY;
    global.lmaxx = global.lmaxy = global.lmaxz = INFINITY;
    return global;
}

void print_boxes(Box *boxes, int n) {
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
* note: distance is from body to octree node COM
*/
int acceptance_test(double cell_size, double distance, double accuracy) {
    return (cell_size / distance) < accuracy;
}

double compute_distance(double dx, double dy, double dz) {
    return sqrt(pow(dx, 2) + pow(dy, 2) + pow(dz, 2) + SOFTENING);
}

void send_body(Body body, int dst_rank) {
    
    return;
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

int main(int argc, char *argv[]) {
    int err;
    err = MPI_Init(&argc, &argv);
    handle_mpi_errcode(err);

    int size; //mum of processes
    int rank;
    MPI_Comm_size(MPI_COMM_WORLD, &size);
    MPI_Comm_rank(MPI_COMM_WORLD, &rank);

    if ((size % 2 != 0 && size != 1) || size > 8) {
        if (rank == 0) {
            fprintf(stderr, "Cannot distribute across an odd number or more than 8 ranks. Aborting all ranks.\n");
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
    if (argc >= 2) num_bodies = atoi(argv[1]);
    if (argc >= 3) num_steps = atoi(argv[2]);

    clock_t start, end;
    if (rank == 0) {
        printf("N-Body Simulation\n");
        printf("-----------------\n");
        printf("Bodies: %d\n", num_bodies);
        printf("Steps : %d\n\n", num_steps);

        start = clock();
    }

    Box global = init_global_space(INITIAL_WIDTH);
    Box *global_octants = divide_box_into_octants(global);
    
    // calc how body ownership will be distributed across ranks
    // each rank is responsible for bodies [starts[r], ends[r])
    int starts[size]; // inclusive
    int ends[size]; // exclusive
    get_idxs_for_each_rank(starts, ends, 8, size); // 8 octants
    printf("Rank %d has octants [%d, %d)\n", rank, starts[rank], ends[rank]);

    // get the parent 3D space for this rank's local octree
    Box rank_parent_boxes[size];
    for (int r = 0; r < size; r++) {
        rank_parent_boxes[r] = merge_boxes(global_octants + starts[r], ends[r] - starts[r]);
    }
    Box local_parent_box = rank_parent_boxes[rank];

    printf("\nRank %d Local Parent: \n", rank);
    print_boxes(&local_parent_box, 1);
    
    Body *bodies;
    int ranks_num_local_bodies[size];
    int num_local_bodies;
    Body *local_bodies;
    if (rank == 0) {
        bodies = (Body *)calloc(num_bodies, sizeof(Body));
        
        if (bodies == NULL) {
            fprintf(stderr, "Error allocating memory.\n");
            MPI_Finalize();
            return EXIT_FAILURE;
        }
        
        srand(0);
        initialize_bodies(bodies, num_bodies);
        print_bodies(bodies, num_bodies);

        // first make array of which rank each body will be sent to
        // and how many bodies each rank will get
        int bodies_dsts[num_bodies];
        int ranks_num_bodies_inserted[size];
        for (int r = 0; r < size; r++) {
            ranks_num_local_bodies[r] = ranks_num_bodies_inserted[r] = 0;
        }
        for (int i = 0; i < num_bodies; i++) {
            for (int r = 0; r < size; r++) {
                if (body_is_inside_box(rank_parent_boxes[r], bodies + i)) {
                    bodies_dsts[i] = r;
                    ranks_num_local_bodies[r]++;
                    break;
                }
            }
        }

        // scatter ranks_num_local_bodies so each rank can alloc local_bodies
        err = MPI_Scatter(
            ranks_num_local_bodies, 1, MPI_INT,
            &num_local_bodies, 1, MPI_INT, 0, MPI_COMM_WORLD
        );
        if (err != MPI_SUCCESS) handle_mpi_errcode(err);

        // now assemble local_body arrays to be scattered to each rank
        Body* ranks_local_bodies[size];
        for (int r = 0; r < size; r++) {
            ranks_local_bodies[r] = (Body *)calloc(ranks_num_local_bodies[r], sizeof(Body));
        }
        int dst_rank;
        for (int i = 0; i < num_bodies; i++) {
            dst_rank = bodies_dsts[i];
            ranks_local_bodies[dst_rank][ranks_num_bodies_inserted[dst_rank]++] = bodies[i];
        }
        
        // send each rank's initial local bodies
        local_bodies = ranks_local_bodies[0];
        for (int r = 1; r < size; r++) {
            MPI_Send(
                ranks_local_bodies[r],
                sizeof(Body)*ranks_num_local_bodies[r],
                MPI_BYTE, r, 0, MPI_COMM_WORLD
            );
            free(ranks_local_bodies[r]);
        }
    } else {
        // get the local bodies for the other ranks
        err = MPI_Scatter(
            ranks_num_local_bodies, 1, MPI_INT,
            &num_local_bodies, 1, MPI_INT, 0, MPI_COMM_WORLD
        );
        if (err != MPI_SUCCESS) handle_mpi_errcode(err);

        // alloc the recv buffer for local bodies
        local_bodies = (Body *)calloc(num_local_bodies, sizeof(Body));
        if (num_local_bodies > 0 && local_bodies == NULL) {
            fprintf(stderr, "Error allocating memory for local_bodies.\n");
            MPI_Finalize();
            return EXIT_FAILURE;
        }

        // recv local bodies from root rank
        MPI_Status status;
        err = MPI_Recv(
            local_bodies, sizeof(Body) * num_local_bodies,
            MPI_BYTE, 0, 0, MPI_COMM_WORLD, &status
        );
        if (err != MPI_SUCCESS) handle_mpi_errcode(err);
    }
    
    OctreeNode octrees[size];
    for (int step = 0; step < num_steps; step++) {
        // recalculate local octree
        OctreeNode *local_octree = build_local_octree(local_parent_box, local_bodies, num_local_bodies);
        // printOctree(local_octree, 0);

        // every rank gets each others' local parent octree
        err = MPI_Allgather(
            local_octree, sizeof(OctreeNode), MPI_BYTE,
            octrees, size * sizeof(OctreeNode), MPI_BYTE, MPI_COMM_WORLD
        );

        compute_forces_dist(local_bodies, num_local_bodies, local_octree, octrees, rank, size);

        update_bodies(bodies, num_bodies, DT, INITIAL_WIDTH);
        
        int num_local_bodies_removed = 0;
        for (int i = 0; i < num_local_bodies; i++) {
            if (body_is_inside_box(local_parent_box, local_bodies + i) == 0) {
                //TODO: find which rank to send the body to
                // MPI_Send(local_bodies[i]);
            }
        }
        
        //TODO: now do an all-to-all where each rank tells the other ranks
        // how many bodies they're sending said rank. Then do send/recv loops
        // based on those counts
        
        err = MPI_Barrier(MPI_COMM_WORLD);
        handle_mpi_errcode(err);

        /*
         * Uncomment for debugging.
         *
         * printf("Step %d\n", step);
         * print_bodies(bodies, num_bodies);
         */
    }

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
    free(global_octants);
    free(local_bodies);

    printf("rank %d success", rank);
    MPI_Finalize();

    return EXIT_SUCCESS;
}