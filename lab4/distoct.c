#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <inttypes.h>
#include <math.h>
#include <time.h>
#include <mpi.h>
#include <string.h>
#include <pthread.h>

#define G 6.67430e-11
#define SOFTENING 1e-1  // was originally 1e-9 but this was negligible and causing numerical "slingshotting" for bodies close together
#define DT 0.01
#define MASS_ORDER 1.0e15
#define ACCURACY 0.5  // google said this was the "sweet spot", though [0.5, 1] is a normal range
#define INITIAL_WIDTH 100.0
#define MAX_DEPTH 18 // because 63 bits can max store 9.2e18 and the way we store node ids

// OctreeNode children statuses for compute_forces_dist
#define NO_CHILDREN 3
#define CHILDREN_NOT_RECEIVED 4
#define CHILDREN_RECEIVED 5

// Inter-rank octree communication
#define TAG_NODE_REQUEST 1
#define TAG_NODE_RESPONSE 2
#define NORMAL_SERVICE 0
#define STOP_SERVICE -1
#define DONT_TRANSFER_BODY -1

// currently only intended for up to 8 ranks at once due to the octree

typedef struct {
    int id; // for comparing start vs. end state
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
    int64_t id;
    int rank; // what rank's local octree this node belongs to
    double mass; // sum of bodies' masses
    double COMx, COMy, COMz; // weighted COM position
    double max_dim;
    Box bbox; // node bounding box
    int childrenStatus;
    struct OctreeNode **childNodes; // NULL if this OctreeNode is a leaf node
} OctreeNode;

typedef struct ServArgs {
    int rank;
    int size;
    OctreeNode* node;
} ServArgs;

typedef struct ServReq {
    int command;
    int64_t id;
    Box bbox;
} ServReq;

/* Function prototypes */
void initialize_bodies(Body *bodies, int n);
void update_bodies(Body *bodies, int n, double dt, int width);
void print_bodies(Body *bodies, int n);

void compute_forces_dist(Body *bodies, int n, OctreeNode *local_octree, OctreeNode **octrees, int rank, int size);
void *service_requests(void *args);

int acceptance_test(double cell_size, double distance, double accuracy);
void get_idxs_for_each_rank(int *starts, int *ends, int n, int size);
void handle_mpi_errcode(int errcode);
int body_is_inside_box(Box box, Body* body);
OctreeNode* build_local_octree(Box box, Body* bodies, int num_bodies, int depth, int64_t id);
OctreeNode *get_node(OctreeNode *node, int64_t id);
void printOctree(OctreeNode *node, int depth);
Box init_global_space(double width);
Box* divide_box_into_octants(Box parent);
Box merge_boxes(Box *boxes, int num_boxes);
double get_box_max_dim(Box box);
int int_pow(int base, int power);
int compare_bodies_by_id(const void *a, const void *b);

/*
 * Initialize bodies with random positions, velocities, and masses.
 */
void initialize_bodies(Body *bodies, int n) {
    for (int i = 0; i < n; i++) {
        bodies[i].id = i;

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
            bodies[i].id,
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

    printf("Node (id = %" PRId64 "):\n", node->id);

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
* Only works for power
*/
int int_pow(int base, int power) {
    if (power < 0) {
        fprintf(stderr, "int_pow only supports positive exponents, %d was input\n", power);
        MPI_Abort(MPI_COMM_WORLD, 1);
        return -1;
    } else if (power == 0) {
        return 1;
    } else {
        int val = base;
        for (int i = 1; i < power; i++) {
            val *= base;
        }
        return val;
    }
}

/*
* TODO: adjust so only bodies that are within the box
* should be passed in for bodies when recursing to the next depth
*/
OctreeNode* build_local_octree(Box box, Body* bodies, int num_bodies, int depth, int64_t id) {
    OctreeNode *node = (OctreeNode *)calloc(1, sizeof(OctreeNode));
    node->bbox = box;
    node->max_dim = get_box_max_dim(box);
    node->id = id;
    MPI_Comm_rank(MPI_COMM_WORLD, &(node->rank));

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
    if ((bodies_in_node > 1) && (depth < MAX_DEPTH)) {
        Box *child_boxes = divide_box_into_octants(box);
        node->childNodes = (OctreeNode **)malloc(8 * sizeof(OctreeNode *));
        for (int i = 0; i < 8; i++) {
            // add another digit to the parent id. tree root is the least-sig-digit
            int64_t child_id = id + (i+1)*int_pow(10, depth);
            node->childNodes[i] = build_local_octree(child_boxes[i], bodies, num_bodies, depth+1, child_id);
        }
        node->childrenStatus = CHILDREN_NOT_RECEIVED; // for inter-rank sending octrees in compute_forces_dist
        free(child_boxes);
    } else {
        if (depth >= MAX_DEPTH) {
            fprintf(stderr, "WARNING, MAX DEPTH REACHED WHILE THERE WERE STILL %d BODIES IN THE NODE\n", bodies_in_node);
        }
        node->childrenStatus = NO_CHILDREN;
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
        MPI_Abort(MPI_COMM_WORLD, 1);
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

int box_eq(Box a, Box b) {
    double tol = 1e-6; // guess for what works
    if (abs(a.minx - b.minx) > tol) return 0;
    if (abs(a.miny - b.miny) > tol) return 0;
    if (abs(a.minz - b.minz) > tol) return 0;

    if (abs(a.maxx - b.maxx) > tol) return 0;
    if (abs(a.maxy - b.maxy) > tol) return 0;
    if (abs(a.maxz - b.maxz) > tol) return 0;

    return 1; // ignore box lims
}

// Comparator for qsort for sorting bodies by id (ascending)
int compare_bodies_by_id(const void *a, const void *b) {
    const Body *bodyA = (const Body *)a;
    const Body *bodyB = (const Body *)b;

    return bodyA->id - bodyB->id;
}

/*
* traverse the tree using least-sig-digit of tree as root
*/
OctreeNode *get_node(OctreeNode *node, int64_t id) {
    if (id < 0) {
        fprintf(stderr, "Node id (%" PRId64 ") < 0  was passed into get_node is required\n", id);
        MPI_Abort(MPI_COMM_WORLD, 1);
        return NULL;
    }
    
    int digit;
    int depth = 0;
    int64_t remaining = id;
    while (remaining > 0 && depth < MAX_DEPTH) {
        if (node->childNodes == NULL) {
            fprintf(stderr, "Node id (%" PRId64 ") was not found!\n", id);
            MPI_Abort(MPI_COMM_WORLD, 1);
            return NULL;
        }

        depth++;
        digit = remaining % 10;
        remaining /= 10; // int div drops remainder

        node = node->childNodes[digit-1]; // digit-1 since id digits are 1-indexed so 0 can be the sentinel
    }
    if (node->id != id) {
        fprintf(stderr, "Node id %" PRId64 " was requested but node id %" PRId64 " was incorrectly returned instead!\n", id, node->id);
        MPI_Abort(MPI_COMM_WORLD, 1);
        return NULL;
    }
    return node;
}


/*
* Recursive helper function for adjusting the accel of one body due to one octree node
*/
void compute_force_on_body_due_to_node(Body *body, OctreeNode *node, int depth) {
    int rank;
    MPI_Comm_rank(MPI_COMM_WORLD, &rank);

    double dx = node->COMx - body->x;
    double dy = node->COMy - body->y;
    double dz = node->COMz - body->z;
    double distance_squared = dx * dx + dy * dy + dz * dz + SOFTENING;
    double distance = sqrt(distance_squared);

    if ((acceptance_test(node->max_dim, distance, ACCURACY) == 1)
    || (node->childrenStatus == NO_CHILDREN)) {
        
        // the acceleration of body i due to node (treated as one body)
        double acceleration = G * node->mass / distance_squared;

        body->ax += acceleration * dx / distance;
        body->ay += acceleration * dy / distance;
        body->az += acceleration * dz / distance;

    } else { // have to step one layer deeper into octree
        // request node from other rank if not yet cached
        if ((node->childrenStatus == CHILDREN_NOT_RECEIVED)
        && (node->rank != rank)) { // if node->rank == rank, we already have the whole tree
            
            node->childNodes = (OctreeNode **)malloc(8 * sizeof(OctreeNode *));
            Box *boxes_to_verify_nodes = divide_box_into_octants(node->bbox);
            for (int i = 0; i < 8; i++) {
                node->childNodes[i] = (OctreeNode *)calloc(1, sizeof(OctreeNode));
                if (node->childNodes[i] == NULL) {
                    fprintf(stderr, "{Rank %d} - Error allocating child node for caching node from rank %d.\n", rank, node->rank);
                    MPI_Abort(MPI_COMM_WORLD, 1);
                }
                int64_t child_id_to_req = node->id + (i+1)*int_pow(10, depth);
                ServReq req = {
                    .command = NORMAL_SERVICE,
                    .id = child_id_to_req,
                    .bbox = boxes_to_verify_nodes[i]
                };
                MPI_Send(
                    &req, sizeof(ServReq), MPI_BYTE,
                    node->rank, TAG_NODE_REQUEST,
                    MPI_COMM_WORLD
                );
                MPI_Recv(
                    node->childNodes[i], sizeof(OctreeNode), MPI_BYTE,
                    node->rank, TAG_NODE_RESPONSE,
                    MPI_COMM_WORLD, MPI_STATUS_IGNORE
                );
            
                compute_force_on_body_due_to_node(body, node->childNodes[i], depth+1);
            }
            node->childrenStatus = CHILDREN_RECEIVED;
            free(boxes_to_verify_nodes);
        } else { // already have children cached
            for (int i = 0; i < 8; i++) {
                compute_force_on_body_due_to_node(body, node->childNodes[i], depth+1);
            }
        }
    }
}

void compute_forces_dist(Body *bodies, int n, OctreeNode *local_octree, OctreeNode **octrees, int rank, int size) {
    /* Reset acceleration */
    for (int i = 0; i < n; i++) {
        bodies[i].ax = 0.0;
        bodies[i].ay = 0.0;
        bodies[i].az = 0.0;
    }
    
    // for each body, calculate the forces/accels due to all octrees
    OctreeNode *node;
    for (int i = 0; i < n; i++) {
        // for each body, go through all octrees (resursively as needed) to calc the forces on it --> its accels
        for (int j = 0; j < size; j++) {
            int r = (rank + j) % size; // so all ranks are staggered for which octree to request from, i.e. so not all ranks are requesting rank 0 for octree info, then all rank 1, etc.
            
            // use the already-fully-cached local octree if it's the current rank rather than rebuilding
            if (r == rank) node = local_octree;
            else           node = octrees[r];
            
            compute_force_on_body_due_to_node(bodies+i, node, 0);
        }
    }
    // free all cached non-local octrees
    for (int r = 0; r < size; r++) {
        free_octree(octrees[r]); // might need to switch octrees to double pointer for this to work
    }
}

void *service_requests(void *arg) {
    ServArgs *args = (ServArgs *)arg;
    OctreeNode *tree = args->node;
    
    MPI_Status status;
    ServReq req;
    OctreeNode *node;

    // service loop ends when thread receives shutdown msg from same rank
    while(1) {
        MPI_Recv(
            &req, sizeof(ServReq), MPI_BYTE,
            MPI_ANY_SOURCE, TAG_NODE_REQUEST,
            MPI_COMM_WORLD, &status
        );

        if (req.command == STOP_SERVICE) break;

        node = get_node(tree, req.id);
        if (node == NULL) {
            fprintf(stderr, "{Rank %d Service Thread} - Node id (%" PRId64 ") not found!\n", args->rank, req.id);
            fprintf(stderr, "{Rank %d Service Thread} - Local Tree Contents:\n", args->rank);
            printOctree(tree, 0);
            MPI_Abort(MPI_COMM_WORLD, 1);
            return NULL;
        }

        MPI_Send(
            node, sizeof(OctreeNode), MPI_BYTE,
            status.MPI_SOURCE, TAG_NODE_RESPONSE,
            MPI_COMM_WORLD
        );
    }
    return NULL;
}

int main(int argc, char *argv[]) {
    int err, provided;
    err = MPI_Init_thread(&argc, &argv, MPI_THREAD_MULTIPLE, &provided);
    handle_mpi_errcode(err);

    if (provided < MPI_THREAD_MULTIPLE) {
        fprintf(stderr, "MPI_THREAD_MULTIPLE is required\n");
        MPI_Abort(MPI_COMM_WORLD, 1);
    }

    int size; //mum of processes
    int rank;
    MPI_Comm_size(MPI_COMM_WORLD, &size);
    MPI_Comm_rank(MPI_COMM_WORLD, &rank);

    if ((size % 2 != 0 && size != 1) || size > 8) {
        if (rank == 0) {
            fprintf(stderr, "Cannot distribute across an odd number or more than 8 ranks. Aborting all ranks.\n");
        }
        MPI_Abort(MPI_COMM_WORLD, 1);
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
            MPI_Abort(MPI_COMM_WORLD, 1);
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
            err = MPI_Send(
                ranks_local_bodies[r],
                sizeof(Body)*ranks_num_local_bodies[r],
                MPI_BYTE, r, 0, MPI_COMM_WORLD
            );
            if (err != MPI_SUCCESS) handle_mpi_errcode(err);
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
            MPI_Abort(MPI_COMM_WORLD, 1);
        }

        // recv local bodies from root rank
        MPI_Status status;
        err = MPI_Recv(
            local_bodies, sizeof(Body) * num_local_bodies,
            MPI_BYTE, 0, 0, MPI_COMM_WORLD, &status
        );
        if (err != MPI_SUCCESS) handle_mpi_errcode(err);
    }
    
    pthread_t serv_thread;
    OctreeNode octrees[size];
    for (int step = 0; step < num_steps; step++) {
        // recalculate local octree
        OctreeNode *local_octree = build_local_octree(local_parent_box, local_bodies, num_local_bodies, 0, 0);
        // printOctree(local_octree, 0);

        // every rank gets each others' local parent octree
        err = MPI_Allgather(
            local_octree, sizeof(OctreeNode), MPI_BYTE,
            octrees, sizeof(OctreeNode), MPI_BYTE, MPI_COMM_WORLD
        );
        OctreeNode *octree_ptrs[size];
        for (int r = 0; r < size; r++) {
            octree_ptrs[r] = (OctreeNode *)calloc(1, sizeof(OctreeNode)); // do so can use free_octree later
            *(octree_ptrs[r]) = octrees[r]; // copy parent node values
        }

        ServArgs serv_args = {.rank = rank, .size = size, .node = local_octree};
        err = pthread_create(&serv_thread, NULL, service_requests, &serv_args);
        if (err != 0) {
            fprintf(stderr, "Rank %d: pthread_create failed: %d\n", rank, err);
            MPI_Abort(MPI_COMM_WORLD, 1);
        }

        compute_forces_dist(local_bodies, num_local_bodies, local_octree, octree_ptrs, rank, size);
        update_bodies(local_bodies, num_local_bodies, DT, INITIAL_WIDTH);
        
        // main calc thread waits for other ranks to finish, while service thread continues servicing
        err = MPI_Barrier(MPI_COMM_WORLD);
        handle_mpi_errcode(err);

        // now that all ranks finished calcs, send service thread stop signal
        ServReq stop = {STOP_SERVICE, -1, {0}};
        MPI_Send(&stop, sizeof(ServReq), MPI_BYTE, rank, TAG_NODE_REQUEST, MPI_COMM_WORLD);

        err = pthread_join(serv_thread, NULL);
        if (err != 0) {
            fprintf(stderr, "pthread_join failed: %d\n", err);
            MPI_Abort(MPI_COMM_WORLD, 1);
        }

        // identify any bodies that moved into a different rank's space
        int ranks_to_transfer_bodies_to[num_local_bodies > 0 ? num_local_bodies : 1];
        for (int i = 0; i < num_local_bodies; i++) {
            ranks_to_transfer_bodies_to[i] = DONT_TRANSFER_BODY;
        }

        int num_bodies_to_send_per_rank[size];
        int num_bodies_to_send_total = 0;
        for (int r = 0; r < size; r++) {
            num_bodies_to_send_per_rank[r] = 0;
        }

        for (int i = 0; i < num_local_bodies; i++) {
            if (!body_is_inside_box(local_parent_box, local_bodies + i)) {
                // find and store which rank to send the body to
                for (int r = 0; r < size; r++) {
                    if (r == rank) continue; // no need to check the current rank
                    if (body_is_inside_box(octrees[r].bbox, local_bodies + i)) {
                        ranks_to_transfer_bodies_to[i] = r;
                        num_bodies_to_send_per_rank[r]++;
                        num_bodies_to_send_total++;
                        break;
                    }
                }
            }
        }
        
        // each rank tells the other ranks how many bodies they're sending them.
        int num_bodies_to_recv_per_rank[size];
        for (int r = 0; r < size; r++) {
            num_bodies_to_recv_per_rank[r] = 0;
        }

        err = MPI_Alltoall(
            num_bodies_to_send_per_rank, 1, MPI_INT,
            num_bodies_to_recv_per_rank, 1, MPI_INT,
            MPI_COMM_WORLD
        );
        if (err != MPI_SUCCESS) handle_mpi_errcode(err);
        
        // make send/recv amnt/disp arrays for Alltoallv
        int num_bodies_to_recv_total = 0;
        int num_bytes_to_send_per_rank[size];
        int send_disp_bodies_per_rank[size];
        int send_disp_bytes_per_rank[size];
        int num_bytes_to_recv_per_rank[size];
        int recv_disp_bytes_per_rank[size];
        for (int r = 0; r < size; r++) {
            num_bytes_to_send_per_rank[r] = 0;
            send_disp_bodies_per_rank[r] = 0;
            send_disp_bytes_per_rank[r] = 0;
            num_bytes_to_recv_per_rank[r] = 0;
            recv_disp_bytes_per_rank[r] = 0;
        }
        
        for (int r = 0; r < size; r++) {
            num_bodies_to_recv_total += num_bodies_to_recv_per_rank[r];
            
            num_bytes_to_send_per_rank[r] = sizeof(Body) * num_bodies_to_send_per_rank[r];
            num_bytes_to_recv_per_rank[r] = sizeof(Body) * num_bodies_to_recv_per_rank[r]; 
            
            if (r < size-1) {
                send_disp_bodies_per_rank[r+1] = send_disp_bodies_per_rank[r] + num_bodies_to_send_per_rank[r]; 
                send_disp_bytes_per_rank[r+1] = send_disp_bytes_per_rank[r] + num_bytes_to_send_per_rank[r]; 
                recv_disp_bytes_per_rank[r+1] = recv_disp_bytes_per_rank[r] + num_bytes_to_recv_per_rank[r]; 
            }        
        }
        
        // make room in recv buff for incoming new bodies if necessary
        Body *tmp;
        if (num_bodies_to_recv_total > 0) {
            tmp = (Body *)realloc(local_bodies, 
                (num_local_bodies + num_bodies_to_recv_total) * sizeof(Body));
            if (tmp != NULL) {
                local_bodies = tmp;
            } else fprintf(stderr, "{Rank %d} - realloc failed when removing transferred local bodies!\n", rank);
        }
            
        // assemble arr of bodies for sending to other ranks
        int num_bodies_inserted_per_rank[size];
        for (int r = 0; r < size; r++) {
            num_bodies_inserted_per_rank[r] = 0;
        }
        Body bodies_to_send[num_bodies_to_send_total > 0 ? num_bodies_to_send_total : 1];
        int idxs_of_sent_bodies[num_bodies_to_send_total > 0 ? num_bodies_to_send_total : 1];
        int j = 0;
        for (int i = 0; i < num_local_bodies; i++) {
            int r = ranks_to_transfer_bodies_to[i];
            if (r == DONT_TRANSFER_BODY) {
                continue;
            }
            int insert_idx = send_disp_bodies_per_rank[r] + num_bodies_inserted_per_rank[r];
            num_bodies_inserted_per_rank[r]++;
            bodies_to_send[insert_idx] = bodies[i];
            idxs_of_sent_bodies[j++] = i;
        }

        // directly recv the new bodies into the bodies arr
        Body *recv_buff = num_bodies_to_recv_total > 0 ? local_bodies + num_local_bodies : NULL;
        
        // Now actually send the bodies to each other (Alltoallv?)
        err = MPI_Alltoallv(
            bodies_to_send,
            num_bytes_to_send_per_rank, send_disp_bytes_per_rank,
            MPI_BYTE,
            recv_buff,
            num_bytes_to_recv_per_rank, recv_disp_bytes_per_rank,
            MPI_BYTE,
            MPI_COMM_WORLD
        );
        if (err != MPI_SUCCESS) handle_mpi_errcode(err);

        // remove the old local_bodies that got sent to dif ranks
        for (int i = 0; i < num_bodies_to_send_total; i++) {
            int idx_to_replace = idxs_of_sent_bodies[i];
            local_bodies[idx_to_replace] = local_bodies[(num_local_bodies-1) - i];
        }

        num_local_bodies = num_local_bodies + num_bodies_to_recv_total 
                                            - num_bodies_to_send_total;

        if (num_local_bodies == 0) {
            free(local_bodies);
            local_bodies = NULL;
        } else if (num_bodies_to_recv_total != 0 || num_bodies_to_send_total != 0) {
            tmp = (Body *)realloc(local_bodies, num_local_bodies * sizeof(Body));
            if (tmp == NULL) {
                fprintf(stderr, "{Rank %d} - realloc failed when removing transferred local bodies!\n", rank);
                MPI_Abort(MPI_COMM_WORLD, 1);
            }
            local_bodies = tmp;
        }

        // free all octree copies since we have to remake them next timestep
        // for (int r = 0; r < size; r++) {
        //     free_octree(octree_ptrs[r]);
        // }

        err = MPI_Barrier(MPI_COMM_WORLD);
        handle_mpi_errcode(err);

        /*
         * Uncomment for debugging.
         *
         * printf("Step %d\n", step);
         * print_bodies(bodies, num_bodies);
         */
    }

    // gather how many bodies each rank has for gatherv
    err = MPI_Gather(
        &num_local_bodies, 1, MPI_INT,
        rank == 0 ? ranks_num_local_bodies : NULL,
        1, MPI_INT, 0, MPI_COMM_WORLD
    );
    if (err != MPI_SUCCESS) handle_mpi_errcode(err);
    
    // make count/disp arrs for gatherv of all ranks' bodies
    int ranks_bodies_count_bytes[size];
    int ranks_bodies_disp_bytes[size];
    if (rank == 0) {
        for (int r = 0; r < size; r++) {
            ranks_bodies_count_bytes[r] = 0;
            ranks_bodies_disp_bytes[r] = 0;
        }

        for (int r = 0; r < size; r++) {
            ranks_bodies_count_bytes[r] = sizeof(Body) * ranks_num_local_bodies[r];
            if (r < size-1) {
                ranks_bodies_disp_bytes[r+1] = ranks_bodies_disp_bytes[r] + ranks_bodies_count_bytes[r]; 
            }        
        }
    }
    
    // gather all the bodies back to root
    err = MPI_Gatherv(
        local_bodies, num_local_bodies * sizeof(Body), MPI_BYTE,
        rank == 0 ? bodies: NULL,
        rank == 0 ? ranks_bodies_count_bytes: NULL,
        rank == 0 ? ranks_bodies_disp_bytes: NULL,
        MPI_BYTE, 0, MPI_COMM_WORLD
    );
    if (err != MPI_SUCCESS) handle_mpi_errcode(err);

    // sort all the bodies by id so that they print in the original order
    if (rank == 0) {
        qsort(bodies, num_bodies, sizeof(Body), compare_bodies_by_id);
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