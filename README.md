# rdtagent
a daemon for Intel rdt

cos usage example:

Assume you have stared your server on http://127.0.0.1:8888

Query cpuinfo of host

	curl -i http://127.0.0.1:8888/v1/cpuinfo
	curl -i http://127.0.0.1:8888/v1/cpuinfo/topology
	curl -i http://127.0.0.1:8888/v1/cpuinfo/capabilities

Query cos of the host

	curl -H "Content-Type: application/json"  http://127.0.0.1:8888/v1/cache/cos

Set mask=3 (low 2 bit) on cos 1 of socket 0

	curl -H "Content-Type: application/json" --request PUT --data '{"Mask": 3}' http://127.0.0.1:8888/v1/cache/cos/0/1

Query cos just changed

	curl -H "Content-Type: application/json" http://127.0.0.1:8888/v1/cache/cos/0/1

Pin process 71911 to cpu 0-1

	$ taskset -pc 0-1 71911
	pid 71911's current affinity list: 0-87
	pid 71911's new affinity list: 0,1
	$ taskset -p 71911
	pid 71911's current affinity mask: 3

Associate cpu 0, 1 to cos 1

	curl -H "Content-Type: application/json" http://127.0.0.1:8888/v1/cache/cos/cpu/0
	curl -H "Content-Type: application/json" http://127.0.0.1:8888/v1/cache/cos/cpu/1
	
	curl -H "Content-Type: application/json" --request PUT --data '{"Cos_id": 1}' http://127.0.0.1:8888/v1/cache/cos/cpu/0
	curl -H "Content-Type: application/json" --request PUT --data '{"Cos_id": 1}' http://127.0.0.1:8888/v1/cache/cos/cpu/1
