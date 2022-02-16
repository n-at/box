package main

import (
	"box/dumper"
	"box/notifier"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var (
	globalConfiguration       dumper.GlobalConfiguration
	dumpConfiguration         []dumper.Configuration
	notificationConfiguration notifier.Configuration
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("unable to read config file: %s", err)
	}
	globalConfiguration = dumper.GlobalConfiguration{
		Path:                 "dumps",
		PgdumpExecutable:     "pg_dump",
		Mongodump5Executable: "/mongodb5/bin/mongodump",
		Mongodump4Executable: "/mongodb4/bin/mongodump",
		GbakExecutable:       "/opt/firebird/bin/gbak",
	}
	if err := viper.UnmarshalKey("global", &globalConfiguration); err != nil {
		log.Fatalf("unable to read global configuration: %s", err)
	}
	if err := viper.UnmarshalKey("dumps", &dumpConfiguration); err != nil {
		log.Fatalf("unable to read dumps configuration: %s", err)
	}
	if err := viper.UnmarshalKey("notification", &notificationConfiguration); err != nil {
		log.Fatalf("unable to read notification configuration: %s", err)
	}
}

func main() {
	n := notifier.Notifier{
		Configuration: notificationConfiguration,
	}

	for _, configuration := range dumpConfiguration {
		var d dumper.Dumper
		var err error

		switch configuration.Type {
		case dumper.TypePostgres:
			d, err = dumper.NewPostgres(globalConfiguration, configuration)
		case dumper.TypeMongo:
			d, err = dumper.NewMongo5(globalConfiguration, configuration)
		case dumper.TypeMongoLegacy:
			d, err = dumper.NewMongo4(globalConfiguration, configuration)
		case dumper.TypeFirebirdLegacy:
			d, err = dumper.NewFirebirdLegacy(globalConfiguration, configuration)
		default:
			err = errors.New("unknown dumper type")
		}
		if err != nil {
			log.Errorf("unable to create dumper: %s", err)
		}

		n.Notify(notifier.StatusInfo, configuration.Name, "starting dump")

		if err := d.Dump(); err != nil {
			n.Notify(notifier.StatusError, configuration.Name, err.Error())
		} else {
			n.Notify(notifier.StatusSuccess, configuration.Name, "dump done")
		}
	}
}
