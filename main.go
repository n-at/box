package main

import (
	"box/configuration"
	"box/dumper"
	"box/notifier"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	config, err := configuration.Read("application.yml")
	if err != nil {
		log.Fatalf("unable to read configuration: %s", err)
		return
	}

	n := notifier.Notifier{
		Configuration: config.Notification,
	}

	if err := ensureDirectoryExists(config.Global.Path); err != nil {
		log.Fatalf("unable to create dump path: %s", err)
		return
	}
	if err := ensureDirectoryExists(config.Global.TmpPath); err != nil {
		log.Fatalf("unable to create tmp dump path: %s", err)
		return
	}

	dumpsFilter := make(map[string]bool)
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		dumpsFilter[arg] = true
	}

	for _, dump := range config.Dumps {
		if len(dumpsFilter) > 0 && !dumpsFilter[dump.Name] {
			continue
		}

		var d dumper.Dumper
		var err error

		log.Infof("%s (%s), latest: %v, daily: %v, weekly: %v, monthly: %v",
			dump.Name, dump.Type, dump.Latest, dump.Daily, dump.Weekly, dump.Monthly)

		switch dump.Type {
		case dumper.TypePostgres:
			d, err = dumper.NewPostgres(config.Global, dump)
		case dumper.TypeMongo:
			d, err = dumper.NewMongo5(config.Global, dump)
		case dumper.TypeMongoLegacy:
			d, err = dumper.NewMongo4(config.Global, dump)
		case dumper.TypeFirebirdLegacy:
			d, err = dumper.NewFirebirdLegacy(config.Global, dump)
		case dumper.TypeMysql:
			d, err = dumper.NewMysql(config.Global, dump)
		case dumper.TypeTar:
			d, err = dumper.NewTar(config.Global, dump)
		default:
			err = errors.New("unknown dumper type")
		}
		if err != nil {
			log.Errorf("%s (%s) unable to create dumper: %s", dump.Name, dump.Type, err)
			continue
		}

		n.Notify(notifier.StatusInfo, dump.Name, "starting dump")

		if err := d.Dump(); err != nil {
			log.Errorf("%s (%s) dump error: %s", dump.Name, dump.Type, err)
			n.Notify(notifier.StatusError, dump.Name, err.Error())
		} else {
			log.Infof("%s (%s) dump done", dump.Name, dump.Type)
			n.Notify(notifier.StatusSuccess, dump.Name, "dump done")
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
