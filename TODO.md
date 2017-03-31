# Short term goal

1. discovey hardware info
    a) cpu info, including core(libpqos calls core, but it's cpu) number, l2 cache info l3 cache info
    b) cpu topology, socket->cores->cpu
    c) rdt feature exposed, l2, l3(cdp), cmt(what event)

2. query cache usage/free on each cache resource, this is will only for non-shared cache usage/free
    a) l3 is per socket resource, cache information shoule be structed, we need to know on how much cache left on which socket
    b) l2, to be done

3. cache allocation for a process, a group process
    a) how to pass these process id, I prefer in POST body
    b) will RDTagent handle taskset (cpupinning) ?
    c) libpqos by default prefer set COS on cores/cpu instead of process, need to figure out how libqpos handle this

4. support raw COS passing
    a) libqpos hard code that cpu support COS0-COS16 (total) for each socket, need to think about what's the REST API looks like.
