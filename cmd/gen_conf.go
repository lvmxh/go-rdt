package main

import (
	"flag" // flag is enough for us.
	"fmt"
	"os"
	"strings"

	"openstackcore-rdtagent/cmd/template"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/test/integration/test_helpers" // TODO: should move it to source path
)

var gopherType string

const (
	confPath    = "./rmd.toml"
	defPlatform = "Broadwell"
)

func genDefaultPlatForm() string {
	dpf := strings.Title(cpu.GetMicroArch(cpu.GetSignature()))
	if dpf == "" {
		dpf = defPlatform
	}
	return dpf
}

func genPlatFormMap() map[string]bool {
	m := cpu.NewCPUMap()
	pfm := map[string]bool{}
	for _, value := range m {
		pfm[strings.Title(value)] = true
	}
	return pfm
}

func genPlatFormList(pfm map[string]bool) []string {
	pfs := []string{}
	for k, _ := range pfm {
		pfs = append(pfs, k)
	}
	return pfs
}

func mergeOptions(options ...map[string]interface{}) map[string]interface{} {
	union := make(map[string]interface{})
	for _, o := range options {
		for k, v := range o {
			union[k] = v
		}
	}
	return union
}

func main() {
	path := flag.String("path", confPath, "the path of the generated rmd config.")

	dpf := genDefaultPlatForm()
	pfm := genPlatFormMap()
	pfs := genPlatFormList(pfm)
	platform := flag.String("platform", dpf,
		"the platform than rmd will run, Support PlatForm:\n\t    "+strings.Join(pfs, ", "))
	flag.Parse()

	if _, ok := pfm[*platform]; !ok {
		fmt.Println("Error, unsupport platform:", *platform)
		os.Exit(1)
	}

	//  Skylake, Kaby Lake, Broadwell
	var option = template.Options
	// FIXME hard code, a smart way to load the platform and other optionts automatically.
	if *platform == "Broadwell" {
		option = mergeOptions(template.Options, template.Broadwell)
	}
	if *platform == "Skylake" {
		option = mergeOptions(template.Options, template.Skylake)
	}

	conf, err := testhelpers.FormatByKey(template.Templ, option)
	if err != nil {
		fmt.Println("Error, to generate config file:", err)
		os.Exit(1)
	}

	f, err := os.Create(*path)
	if err != nil {
		fmt.Println("Error, to create config file:", err)
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.WriteString(conf)
	if err != nil {
		fmt.Println("Error, to write config file:", err)
		os.Exit(1)
	}
	f.Sync()
}
