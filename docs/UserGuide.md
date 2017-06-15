# COS usage examples:

(e.g. starting the service on: http://127.0.0.1:8888)

## Query cache information on the host

    curl -i http://127.0.0.1:8888/v1/cache/
    curl -i http://127.0.0.1:8888/v1/cache/l3
    curl -i http://127.0.0.1:8888/v1/cache/l3/0

## Query pre-fedefined policy in RDTAgent

    curl http://127.0.0.1:8888/v1/policy

The backend for policy is a yaml file: /etc/rdtagent/policy.yaml, it pre-defines
some `policy` on specific intel platform. This file can be changed.

P.S. Now, only the peakusage is used, it is the cache we allocated, unit is KiB.

## Create a workload

A workload could be a running task(s) or some cpus which want to allocate
cache for them.

The task(s)/cpus should be valided, RDTAgent will fail your request they
are invalid.

Besides you need to specify what policy of the workload will be usage.

The post body could contains:
{
    "task_ids": A validate task id list
    "core_ids": cpu core list, for the topology, check cache information
    "policy": pre-defined policy in RDTAgent
    "group": optional, multiple workload can join into one group, it's a string list
}

An example:

1) Create a workload with gold policy and specify the group name is "base".

curl -H "Content-Type: application/json" --request POST --data '{"task_ids":["78377"], "policy": "gold", "group": ["base"]}' http://127.0.0.1:8888/v1/workloads

This will result a new resource group created

    root@s2600wt:/sys/fs# cat resctrl/base/schemata
    L3:0=1f;1=1f
    root@s2600wt:/sys/fs# cat resctrl/base/tasks
    78377

2) Create a new workload with silver policy and specify the group name
   ["base", "sub"], the workload want to join "base" (as it is alreay there),
   and its own group name is "sub"

curl -H "Content-Type: application/json" --request POST --data '{"task_ids":["5743"], "policy": "silver", "group": ["base", "sub"]}' http://127.0.0.1:8888/v1/workloads

This will result a new resource being breated, workload 1 and workload 2 shares overlap cache

    root@s2600wt:/sys/fs# cat resctrl/base-sub/schemata
    L3:0=7;1=7
    root@s2600wt:/sys/fs# cat resctrl/base-sub/tasks
    5743

3) Create another workload without group specify

curl -H "Content-Type: application/json" --request POST --data '{"task_ids":["14988"], "policy": "silver", "group": []}' http://127.0.0.1:8888/v1/workloads

This will create another resource group which has isolated cache allocation.

    root@s2600wt:/sys/fs# cat resctrl/14988/schemata
    L3:0=e0;1=e0
    root@s2600wt:/sys/fs# cat resctrl/14988/tasks
    14988

After these the default schemata is:

    cat resctrl/schemata
    L3:0=fff00;1=fff00
