package dumper

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Mongo4Dumper struct {
	AbstractDumper
}

func NewMongo4(global GlobalConfiguration, local Configuration) (*Mongo4Dumper, error) {
	if len(global.Mongodump4Executable) == 0 {
		return nil, errors.New("mongodump executable not found")
	}

	dumper := Mongo4Dumper{
		AbstractDumper{
			globalConfiguration: global,
			configuration:       local,
			time:                time.Now(),
		},
	}

	return &dumper, nil
}

func (dumper *Mongo4Dumper) Dump() error {
	stringBuilder := strings.Builder{}

	//https://docs.mongodb.com/v4.0/reference/program/mongodump/
	//Compatible with MongoDB 4.0-2.6
	//Example configuration:
	//host: "localhost"
	//port: "27017"
	//username: "admin"
	//password: "******"
	//authenticationDatabase: "admin"
	//db: "users"

	stringBuilder.WriteString(fmt.Sprintf("\"%s\" ", esc(dumper.globalConfiguration.Mongodump4Executable)))
	stringBuilder.WriteString("--verbose ")
	stringBuilder.WriteString(fmt.Sprintf("--archive=\"%s\" ", esc(dumper.tmpDumpFileName())))

	for key, value := range dumper.configuration.Vars {
		if key == "verbose" || key == "archive" || key == "output" {
			continue
		}
		stringBuilder.WriteString(fmt.Sprintf("--%s=\"%s\" ", key, esc(value)))
	}

	return dumper.execute(stringBuilder.String())
}
