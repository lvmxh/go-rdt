package conf

import (
	_ "flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Init() error {
	viper.SetConfigName("rdtagent") // no need to include file extension
	// TODO (Shaohe) consider to introduce Cobra. let Viper work with Cobra.
	conf_dir := pflag.Lookup("conf-dir").Value.String()
	if conf_dir != "" {
		viper.AddConfigPath(conf_dir)
	}
	viper.AddConfigPath("/etc/rdtagent/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/rdtagent")  // call multiple times to add many search paths
	viper.AddConfigPath("./etc/rdtagent/") // set the path of your config file
	err := viper.ReadInConfig()
	if err != nil {
		// NOTE (ShaoHe Feng): only can use fmt.Println, can not use log.
		// For log is not init at this point.
		fmt.Println(err)
		return err
	}
	viper.BindPFlags(pflag.CommandLine)
	return nil
}
