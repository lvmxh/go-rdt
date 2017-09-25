package bootcheck

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"openstackcore-rdtagent/db"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/resctrl"
	"openstackcore-rdtagent/util/acl"
	"openstackcore-rdtagent/util/rdtpool"
)

// SanityCheck before string rmd process
func SanityCheck() {
	pf := cpu.GetMicroArch(cpu.GetSignature())
	if pf == "" {
		msg := "Unknow platform, please update the cpu_map.toml conf file."
		log.Fatal(msg)
	}
	if _, err := acl.NewEnforcer(); err != nil {
		msg := "Error to generate an Enforcer! Reason: " + err.Error()
		log.Fatal(msg)
	}
	cpunum := cpu.HostCpuNum()
	if cpunum == 0 {
		msg := "Unable to get Total CPU numbers on Host."
		log.Fatal(msg)
	}
	if !resctrl.IsIntelRdtMounted() {
		msg := "resctrl does not enable."
		log.Fatal(msg)
	}
	if err := DBCheck(); err != nil {
		msg := "Check db error. Reason: " + err.Error()
		log.Fatal(msg)
	}
	if err := rdtpool.SetOSGroup(); err != nil {
		msg := "Error, create OS groups failed! Reason: " + err.Error()
		log.Fatal(msg)
	}
	if err := rdtpool.SetInfraGroup(); err != nil {
		msg := "Error, create infra groups failed! Reason: " + err.Error()
		log.Fatal(msg)
	}
	v, err := rdtpool.GetCachePoolLayout()
	log.Debugf("Cache Pool layout %v", v)
	if err != nil {
		msg := "Error while get cache pool layout Reason: " + err.Error()
		log.Fatal(msg)
	}
}

// DBCheck Do some cleanup in DB
func DBCheck() error {
	d, err := db.NewDB()
	if err != nil {
		return err
	}

	err = d.Initialize("", "")
	if err != nil {
		return err
	}

	resaall := resctrl.GetResAssociation()

	wl, err := d.GetAllWorkload()
	if err != nil {
		return err
	}

	for _, w := range wl {
		switch w.CosName {
		case "":
			d.DeleteWorkload(&w)
		case "os":
		case "OS":
		case ".":
		case "infra":
			// FIXME Now we can allow to create multi-infra, need clean?
		case "default":
		default:
			if v, ok := resaall[w.CosName]; !ok {
				d.DeleteWorkload(&w)
				fmt.Println(v)
			}
			// FIXME, delete the group with null tasks and zero cpus.
		}
	}
	return nil

}
