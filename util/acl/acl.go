package acl

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/casbin/casbin"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	. "openstackcore-rdtagent/util/acl/config"
)

type Enforcer struct {
	url      *casbin.Enforcer
	ip       *casbin.Enforcer
	protocol *casbin.Enforcer
}

var VERSION_TRIM = regexp.MustCompile(`^/v\d+/`)
var enforcer = &Enforcer{}
var once sync.Once

func NewEnforcer() (*Enforcer, error) {
	var return_err error
	defer func() {
		if r := recover(); r != nil {
			return_err = fmt.Errorf("init Enforcer error: %s\n", r)
		}
	}()

	once.Do(func() {
		aclconf := NewACLConfig()
		for _, filter := range strings.Split(aclconf.Filter, ",") {
			model := path.Join(aclconf.Path, filter, "model.conf")
			policy := path.Join(aclconf.Path, filter, "policy.csv")
			switch filter {
			case "url":
				enforcer.url = casbin.NewEnforcer(model, policy)
				log.Infof("succssfully set %s acl", filter)
			case "ip":
				log.Infof("Do not support %s acl at present", filter)
			case "proto":
				log.Infof("Do not support %s acl at present", filter)
			default:
				log.Errorf("Unknow acl type %s", filter)
			}

		}
	})
	return enforcer, return_err
}

func (e *Enforcer) Enforce(request *restful.Request, sub string) bool {
	allow := false
	obj := VERSION_TRIM.ReplaceAllString(path.Clean(request.Request.RequestURI), "/")
	act := request.Request.Method
	if e.url != nil {
		allow = e.url.Enforce(sub, obj, act)
	} else {
		allow = true
	}
	// TODO support ip and proto Enforce.
	return allow
}
