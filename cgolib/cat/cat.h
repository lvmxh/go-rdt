struct cgo_cos {
    unsigned socket_id;
    unsigned cos_id;
    unsigned long long mask;
};
const struct pqos_cpuinfo *cgo_cat_init();
const struct cgo_cos *cgo_cat_get_cos(unsigned, unsigned*);
const struct cgo_cos *cgo_cat_set_cos(unsigned, unsigned, unsigned*, unsigned long long);
