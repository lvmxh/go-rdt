package infragroup

import (
	_ "fmt"
	_ "strconv"
	_ "sync"

	_ "openstackcore-rdtagent/lib/cache"
	_ "openstackcore-rdtagent/lib/resctrl"
	// util "openstackcore-rdtagent/lib/util"
	. "openstackcore-rdtagent/util/bootcheck/infragroup/config"
	_ "openstackcore-rdtagent/util/rdtpool/base"
)

func SetInfraGroup() error {
	conf := NewConfig()
	if conf == nil {
		return nil
	}
	return nil
}
