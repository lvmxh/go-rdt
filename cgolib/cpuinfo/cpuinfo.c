#include <cpuinfo.h>
#include <pqos.h>
#include <stdio.h>
#include <log.h>
#include <string.h>

const struct pqos_cpuinfo * cgo_cpuinfo_init()
{
    struct pqos_config cfg;
    int ret;
    const struct pqos_cpuinfo *m_cpu = NULL;
    memset(&cfg, 0, sizeof(cfg));
        cfg.fd_log = STDOUT_FILENO;
        cfg.verbose = 0;

    ret = log_init(cfg.fd_log,
        cfg.callback_log,
        cfg.context_log,
        cfg.verbose);

    if (ret != LOG_RETVAL_OK) {
        fprintf(stderr, "log_init() error\n");
        return NULL;
    }

    cpuinfo_init(&m_cpu);
    printf("get cpuinfo successfully. Total %d cores.\n", m_cpu->num_cores);
    return m_cpu;
}
