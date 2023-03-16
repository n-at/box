package configuration

import (
	"box/dumper"
	"box/notifier"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Configuration struct {
	Global       dumper.GlobalConfiguration
	Dumps        []dumper.Configuration
	Notification notifier.Configuration
}

func Read(fileName string) (*Configuration, error) {
	config := Configuration{
		Global: dumper.GlobalConfiguration{
			Path:                 "dumps",
			ShExecutable:         "sh",
			PgdumpExecutable:     "pg_dump",
			MysqldumpExecutable:  "mysqldump",
			Mongodump5Executable: "/mongodb5/bin/mongodump",
			Mongodump4Executable: "/mongodb4/bin/mongodump",
			GbakExecutable:       "/opt/firebird/bin/gbak",
			TarExecutable:        "tar",
		},
		Dumps: []dumper.Configuration{},
		Notification: notifier.Configuration{
			Enabled: false,
		},
	}

	if _, err := os.Stat(fileName); err != nil {
		return nil, fmt.Errorf("configuration file %s not found: %s", fileName, err)
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to open configuration file: %s", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read configuration file: %s", err)
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("unable to parse configuration file: %s", err)
	}

	return &config, nil
}
