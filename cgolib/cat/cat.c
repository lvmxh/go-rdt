#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "pqos.h"
#include "cat.h"

/* init before doing allocation */
const struct pqos_cpuinfo *cgo_cat_init()
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
    return p_cpu;
}

/* Remember to free p_cos in cgo interface */
const struct cgo_cos *cgo_cat_get_cos(unsigned socket, unsigned *num)
{
	/* Get CPU socket information to set COS */
    int ret;
    struct pqos_l3ca tab[PQOS_MAX_L3CA_COS];
    unsigned n = 0;
    *num = 0;
    struct cgo_cos *p_cos = NULL;

    ret = pqos_l3ca_get(socket, PQOS_MAX_L3CA_COS, num, tab);
    if (ret == PQOS_RETVAL_OK) {
        p_cos = (struct cgo_cos*) malloc(sizeof(struct cgo_cos) * (*num));
        for (n = 0; n < *num; n++) {
            p_cos[n].socket_id = socket;
            p_cos[n].cos_id = tab[n].class_id;
            p_cos[n].mask = (unsigned long long)tab[n].u.ways_mask;
        }
    }
    return p_cos;
}

/* Remember to free p_cos in cgo interface */
const struct cgo_cos *cgo_cat_set_cos(unsigned socket, unsigned cos_id,
        unsigned *num,
        unsigned long long mask)
{
	/* Get CPU socket information to set COS */
    int ret;
    struct pqos_l3ca tab[PQOS_MAX_L3CA_COS];
    unsigned n = 0;
    *num = 0;
    struct cgo_cos *p_cos;

    tab[0].class_id = cos_id;
    tab[0].u.ways_mask = mask;

    ret = pqos_l3ca_set(socket, 1, tab);
    if (ret != PQOS_RETVAL_OK) {
        return NULL;
    }
    ret = pqos_l3ca_get(socket, PQOS_MAX_L3CA_COS, num, tab);
    if (ret == PQOS_RETVAL_OK) {
        p_cos = (struct cgo_cos*) malloc(sizeof(struct cgo_cos) * (*num));
        for (n = 0; n < *num; n++) {
            p_cos[n].socket_id = socket;
            p_cos[n].cos_id = tab[n].class_id;
            p_cos[n].mask = (unsigned long long)tab[n].u.ways_mask;
        }
    }
    return p_cos;
}
