package main

import (
	"box/dumper"
	"box/notifier"
	"errors"
	"fmt"
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
		ShExecutable:         "sh",
		PgdumpExecutable:     "pg_dump",
		Mongodump5Executable: "/mongodb5/bin/mongodump",
		Mongodump4Executable: "/mongodb4/bin/mongodump",
		GbakExecutable:       "/opt/firebird/bin/gbak",
		TarExecutable:        "tar",
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

	if err := ensureDirectoryExists(globalConfiguration.Path); err != nil {
		log.Fatalf("unable to create dump path: %s", err)
	}
	if err := ensureDirectoryExists(globalConfiguration.TmpPath); err != nil {
		log.Fatalf("unable to create tmp dump path: %s", err)
	}

	dumpsFilter := make(map[string]bool)
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		dumpsFilter[arg] = true
	}

	for _, configuration := range dumpConfiguration {
		if len(dumpsFilter) > 0 && !dumpsFilter[configuration.Name] {
			continue
		}

		var d dumper.Dumper
		var err error

		log.Infof("%s (%s), daily: %v, weekly: %v, monthly: %v",
			configuration.Name, configuration.Type, configuration.Daily, configuration.Weekly, configuration.Monthly)

		switch configuration.Type {
		case dumper.TypePostgres:
			d, err = dumper.NewPostgres(globalConfiguration, configuration)
		case dumper.TypeMongo:
			d, err = dumper.NewMongo5(globalConfiguration, configuration)
		case dumper.TypeMongoLegacy:
			d, err = dumper.NewMongo4(globalConfiguration, configuration)
		case dumper.TypeFirebirdLegacy:
			d, err = dumper.NewFirebirdLegacy(globalConfiguration, configuration)
		case dumper.TypeTar:
			d, err = dumper.NewTar(globalConfiguration, configuration)
		default:
			err = errors.New("unknown dumper type")
		}
		if err != nil {
			log.Errorf("%s (%s) unable to create dumper: %s", configuration.Name, configuration.Type, err)
			continue
		}

		n.Notify(notifier.StatusInfo, configuration.Name, "starting dump")

		if err := d.Dump(); err != nil {
			log.Errorf("%s (%s) dump error: %s", configuration.Name, configuration.Type, err)
			n.Notify(notifier.StatusError, configuration.Name, err.Error())
		} else {
			log.Infof("%s (%s) dump done", configuration.Name, configuration.Type)
			n.Notify(notifier.StatusSuccess, configuration.Name, "dump done")
		}
	}
}

func ensureDirectoryExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(path, 0777); err != nil {
				return err
			} else {
				return nil
			}
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", path)
	} else {
		return nil
	}
}
