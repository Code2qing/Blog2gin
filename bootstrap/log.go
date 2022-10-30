package bootstrap

import (
	log "github.com/sirupsen/logrus"
)

// InitLog init log
func init() {
	log.SetFormatter(&log.TextFormatter{
		//DisableColors: true,
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	})
	log.Infof("init log...")
}
