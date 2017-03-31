#include <pqos.h>
#include <stdio.h>
#include <log.h>
#include <string.h>

const struct pqos_cpuinfo * cgo_cpuinfo_init()
{
    struct pqos_config cfg;
    int ret;
    const struct pqos_cpuinfo *p_cpu = NULL;
    const struct pqos_cap *p_cap = NULL;
    memset(&cfg, 0, sizeof(cfg));
        cfg.fd_log = STDOUT_FILENO;
        cfg.verbose = 0;

    ret = pqos_init(&cfg);
    if (ret != PQOS_RETVAL_OK) {
        fprintf(stderr, "Error initializing PQos library!\n");
        return NULL;
    }

    ret = pqos_cap_get(&p_cap, &p_cpu);
    if (ret != PQOS_RETVAL_OK) {
        fprintf(stderr, "Error retriveving PQoS cpuinfo!\n");
        return NULL;
    }

    printf("get cpuinfo successfully. Total %d cores.\n", p_cpu->num_cores);
    return p_cpu;
}

/* init cap */
const struct pqos_cap *cgo_cap_init()
{
    struct pqos_config cfg;
    int ret;
    const struct pqos_cpuinfo *p_cpu = NULL;
    const struct pqos_cap *p_cap = NULL;
    memset(&cfg, 0, sizeof(cfg));
        cfg.fd_log = STDOUT_FILENO;
        cfg.verbose = 0;

    ret = pqos_init(&cfg);
    if (ret != PQOS_RETVAL_OK) {
        fprintf(stderr, "Error initializing PQos library!\n");
        return NULL;
    }

    ret = pqos_cap_get(&p_cap, &p_cpu);
    if (ret != PQOS_RETVAL_OK) {
        fprintf(stderr, "Error retriveving PQoS capabilities!\n");
        return NULL;
    }
    printf("caps = %u\n", p_cap->num_cap);
    return p_cap;
}
