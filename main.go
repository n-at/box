package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
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
		log.Fatalf("Unable to read config file: %s", err)
	}
}

func main() {
	log.Info("Hello!")
}
