package main

import (
	"box/dumper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var (
	globalConfiguration dumper.GlobalConfiguration
	dumpConfiguration   []dumper.Configuration
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
		Path:                "dumps",
		PgdumpExecutable:    "pg_dump",
		MongodumpExecutable: "mongodump",
		GbakExecutable:      "gbak",
	}
	if err := viper.UnmarshalKey("global", &globalConfiguration); err != nil {
		log.Fatalf("unable to read global configuration: %s", err)
	}
	if err := viper.UnmarshalKey("dumps", &dumpConfiguration); err != nil {
		log.Fatalf("unable to read dumps configuration: %s", err)
	}
}

func main() {
	log.Info("Hello!")
}
