package log

import (
	"github.com/sirupsen/logrus"
	. "openstackcore-rdtagent/util/log/config"
	"os"
)

// TODO (Shaohe): Need to support Model name and file line fields.
func Init() error {
	config := NewConfig()
	l, _ := logrus.ParseLevel(config.Level)
	// FIXME (shaohe), we do not support both stdout and file at the same time.
	if config.Stdout {
		logrus.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(
			config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0666)
		if err != nil {
			return err
		}
		logrus.SetOutput(f)
	}

	logrus.SetLevel(l)
	if config.Env == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	return nil
}
