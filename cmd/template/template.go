// Package template TODO: This template is too sample. need to improve it.
// Not sure golang native template can satify us.
package template

// Options are options for temlate will be replaced
var Options = map[string]interface{}{
	"address":         "localhost",
	"debugport":       8081,
	"tlsport":         8443,
	"clientauth":      "challenge",
	"logfile":         "/var/log/rdagent.log",
	"dbbackend":       "bolt",
	"dbtransport":     "/var/run/rdtagent.db",
	"logtostdout":     true,
	"os_cacheways":    1,
	"infra_cacheways": 19,
	"max_shared":      10,
	"guarantee":       10,
	"besteffort":      7,
	"shared":          2,
	"shrink":          false}

// Templ is content of template
const Templ = `# This is a rdtagent config.

title = "RDTagent Config"

[default]
address = "{{.address}}"
policypath = "/etc/rdtagent/policy.toml"
tlsport = {{.tlsport}}
# certpath = "/etc/rdtagent/cert/server" # Only support pem format, hard code that CAFile is ca.pem, CertFile is rmd-cert.pem, KeyFile is rmd-key.pem
# clientcapath = "/etc/rdtagent/cert/client" # Only support pem format, hard code that CAFile is ca.pem
clientauth = "{{.clientauth}}"  # can be "no, require, require_any, challenge_given, challenge", challenge means require and verify.
# unixsock = "/var/run/rmd.sock"

[log]
path = "{{.logfile}} "
env = "dev"  # production or dev
level = "debug"
stdout = {{.logtostdout}}

[database]
backend = "{{.dbbackend}}" # mgo
transport = "{{.dbtransport}}" # mongodb://localhost
dbname = "bolt" # RDTPolicy

[debug]
enabled = false
debugport = {{.debugport}}

[OSGroup] # OSGroup is mandatory
cacheways = {{.os_cacheways}}
cpuset = "0-1"

[InfraGroup] # InfraGroup is optional
cacheways = {{.infra_cacheways}}
cpuset = "2-3"
# arrary or comma-separated values? RMD supports array instead of CSV.
tasks = ["ovs*"] # Just support Wildcards. Do we need to support RE?

[CachePool] # Cache Pool config is optional
shrink = {{.shrink}}
max_allowed_shared = {{.max_shared}} # max allowed workload in shared pool, default is {{.max_shared}}
guarantee = {{.guarantee}}
besteffort = {{.besteffort}}
shared = {{.shared}}

[acl]
path = "/etc/rdtagent/acl/"  #
# use CSV format
filter = "url" # at present just support "url", will support "url,ip,proto"
authorization = "role" # authorize the client, can identify client by signature, role(OU) or username(CN). Default value is signature. If value is signature, admincert and usercert should be set.
admincert = "/etc/rdtagent/acl/roles/admin/" # A cert is used to describe user info. These cert files in this path are used to define the users that are admin. Only pem format file at present. The files can be updated dynamicly.
usercert = "/etc/rdtagent/acl/roles/user/" # A cert is used to describe user info. These cert files in this path are used to define the user with low privilege. Only pem format file at present. The files can be updated dynamicly.

[pam]
service = "rmd"
`
