package dumper

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Mongo5Dumper struct {
	AbstractDumper
}

func NewMongo5(global GlobalConfiguration, local Configuration) (*Mongo5Dumper, error) {
	if len(global.Mongodump5Executable) == 0 {
		return nil, errors.New("mongodump executable not found")
	}

	dumper := Mongo5Dumper{
		AbstractDumper{
			globalConfiguration: global,
			configuration:       local,
			time:                time.Now(),
		},
	}

	return &dumper, nil
}

func (dumper *Mongo5Dumper) Dump() error {
	stringBuilder := strings.Builder{}

	//https://docs.mongodb.com/database-tools/mongodump/
	//Compatible with MongoDB 5.0-4.0
	//Example configuration:
	//host: "localhost"
	//port: "27017"
	//username: "admin"
	//password: "******"
	//authenticationDatabase: "admin"
	//db: "users"

	stringBuilder.WriteString(fmt.Sprintf("\"%s\" ", esc(dumper.globalConfiguration.Mongodump5Executable)))
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
