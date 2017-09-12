// TODO: This template is too sample. need to improve it.
// Not sure golang native template can satify us.
package template

var Options = map[string]interface{}{
	"address":         "localhost",
	"tcpport":         8081,
	"os_cacheways":    1,
	"infra_cacheways": 19,
	"max_shared":      10,
	"guarantee":       10,
	"besteffort":      7,
	"shared":          2}

const Templ = `# This is a rdtagent config.

title = "RDTagent Config"

[default]
address = "{{.address}}"
port = {{.tcpport}}
policypath = "/etc/rdtagent/policy.yaml"
# tlsport = 443
# certpath = "etc/rdtagent/cert" # Only support pem format, hard code that CAFile is ca.pem, CertFile is rmd-cert.pem, KeyFile is rmd-key.pem
# clientauth = "challenge"  # can be "noneed, require, require_any, challenge_given, challenge", challenge means require and verify.
# unixsock = "./var/run/rmd.sock"

[log]
path = "/var/log/rdagent.log"
env = "dev"  # production or dev
level = "debug"
stdout = true

[database]
backend = "bolt" # mgo
transport = "/var/run/rdtagent.db" # mongodb://localhost
dbname = "bolt" # RDTPolicy

[OSGroup] # OSGroup is mandatory
cacheways = {{.os_cacheways}}
cpuset = "0-1"

[InfraGroup] # InfraGroup is optional
cacheways = {{.infra_cacheways}}
cpuset = "2-3"
# arrary or comma-separated values? RMD supports array instead of CSV.
tasks = ["ovs*"] # Just support Wildcards. Do we need to support RE?

[CachePool] # Cache Pool config is optional
max_allowed_shared = {{.max_shared}} # max allowed workload in shared pool, default is {{.max_shared}}
guarantee = {{.guarantee}}
besteffort = {{.besteffort}}
shared = {{.shared}}

[acl]
path = "etc/rdtagent/acl/"  #
# use CSV format
filter = "url" # at present just support "url", will support "url,ip,proto"
`
