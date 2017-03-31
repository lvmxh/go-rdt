#include <stdio.h>
#include <string.h>
#include "pqos.h"

struct _pqos_infos {
    const struct pqos_cpuinfo *p_cpu;
    const struct pqos_cap *p_cap;
};

struct _pqos_infos _pqos_init_get()
{
    struct pqos_config cfg;
    int ret;
    struct _pqos_infos infos = {NULL, NULL};

    memset(&cfg, 0, sizeof(cfg));
        cfg.fd_log = STDOUT_FILENO;
        cfg.verbose = 0;

    ret = pqos_init(&cfg);
    if (ret != PQOS_RETVAL_OK) {
        fprintf(stderr, "Error initializing PQos library!\n");
        return infos; /* with NULL values */
    }

    ret = pqos_cap_get(&infos.p_cap, &infos.p_cpu);
    if (ret != PQOS_RETVAL_OK) {
        fprintf(stderr, "Error retriveving PQoS cpuinfo!\n");
    }

    return infos;
}

/* get cpuinfo */
const struct pqos_cpuinfo* cgo_get_cpuinfo()
{
    struct _pqos_infos infos = _pqos_init_get();
    return infos.p_cpu;
}

/* get cap */
const struct pqos_cap* cgo_get_cap()
{
    struct _pqos_infos infos = _pqos_init_get();
    return infos.p_cap;
}
