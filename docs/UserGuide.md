# RMD usage examples:

(e.g. starting the service on: http://127.0.0.1:8888)

## Query cache information on the host

    curl -i http://127.0.0.1:8888/v1/cache/
    curl -i http://127.0.0.1:8888/v1/cache/l3
    curl -i http://127.0.0.1:8888/v1/cache/l3/0

## Query pre-fedefined policy in RDTAgent

    curl http://127.0.0.1:8888/v1/policy

The backend for policy is a yaml file: /etc/rdtagent/policy.yaml, it pre-defines
some `policy` on specific intel platform. This file can be changed.


## Configuration file example

```
[OSGroup] # OSGroup is mandatory
cacheways = 1
cpuset = "0-1"

[InfraGroup] # InfraGroup is optional
cacheways = 19
cpuset = "2-3,24-25"
# arrary or comma-separated values? RMD supports array instead of CSV.
tasks = ["ovs*"] # Just support Wildcards. Do we need to support RE?

[CachePool]
guarantee = 10
besteffort = 9
shared = 7
```

OSGroup: cache ways reserved for operate system usage.
InfraGroup: infrastructure tasks group, user can specify task binary name
            the cache ways will be shared with other worklaod

Besides, define cache pool, cache allocation will happened in the pools.

guarantee: allocate cache for workload max_cache == min_cache > 0
besteffort: allocate cache for workload max_cache > min_cache > 0
shared: allocate cache for workload max_cache == min_cache = 0

On a host which support max 20 cache ways, for this configuration file,
we will have follow cache bit mask layout:

OSGroup:    0000 0000 0000 0000 0001
InfraGroup: 1111 1111 1111 1111 1110

Cache Pools:
guarantee:  0000 0000 0111 1111 1110
besteffort: 1111 1111 1000 0000 0000
shared:     0111 1111 0000 0000 0000


## Create a workload

A workload could be a running task(s) or some cpus which want to allocate
cache for them.

The task(s)/cpus should be valided, RDTAgent will fail your request they
are invalid.

Besides you need to specify what policy of the workload will be usage.

The post body could contains:

```
{
    "task_ids": A validate task id list
    "core_ids": cpu core list, for the topology, check cache information
    "policy": pre-defined policy in RDTAgent
    "max_cache": maximum cache ways which can be benefited
    "min_cache": minmum cache ways which can be benefited
}
```

An example:

1) Create a workload with gold policy.

```
curl -H "Content-Type: application/json" --request POST --data '{"task_ids":["78377"], "policy": "gold"}' http://127.0.0.1:8888/v1/workloads
```

2) Create workload with max_cache, min_cache.

```
curl -H "Content-Type: application/json" --request POST --data '{"task_ids":["14988"], "max_cache": 4, "min_cache": 4}' http://127.0.0.1:8888/v1/workloads
```


## Hospitality score API usage:

Hospitality score API will give a score for scheduling workload on a host for cache allocation request.

Admin can ether give the max_cache/min_cache or a policy to query if the hospitality score.

```
curl -H "Content-Type: application/json" --request POST --data '{"max_cache": 2, "min_cache": 2}' http://127.0.0.1:8888/v1/hospitality
{
    "score": {
        "l3": {
            "0": 50,
            "1": 50
        }
    }
}
```
