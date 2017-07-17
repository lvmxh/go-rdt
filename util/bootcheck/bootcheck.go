package bootcheck

//SanityCheck

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/resctrl"
	"os"
)

func SanityCheck() {
	pf := cpu.GetMicroArch(cpu.GetSignature())
	if pf == "" {
		msg := "Unknow platform, please update the cpu_map.toml conf file."
		log.Fatalf(msg)
		fmt.Println(msg)
		os.Exit(1)
	}
	cpunum := cpu.HostCpuNum()
	if cpunum == 0 {
		msg := "Unable to get Total CPU numbers on Host."
		log.Fatalf(msg)
		fmt.Println(msg)
		os.Exit(1)
	}
	if !resctrl.IsIntelRdtMounted() {
		msg := "resctrl does not enable."
		log.Fatalf(msg)
		fmt.Println(msg)
		os.Exit(1)
	}
}
