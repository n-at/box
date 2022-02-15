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

	outputDirectory := dumper.tmpDumpFileName() + "_dump"

	stringBuilder.WriteString(fmt.Sprintf("\"%s\" ", esc(dumper.globalConfiguration.Mongodump5Executable)))
	stringBuilder.WriteString("--verbose ")
	stringBuilder.WriteString(fmt.Sprintf("--out=\"%s\" ", esc(outputDirectory)))

	for key, value := range dumper.configuration.Vars {
		if key == "verbose" || key == "archive" || key == "out" {
			continue
		}
		stringBuilder.WriteString(fmt.Sprintf("--%s=\"%s\" ", key, esc(value)))
	}

	stringBuilder.WriteString(fmt.Sprintf("&& tar -cvzf \"%s\" --directory \"%s\" . ", esc(dumper.tmpDumpFileName()), esc(outputDirectory)))
	stringBuilder.WriteString(fmt.Sprintf("&& rm --verbose --recursive --force \"%s\" ", esc(outputDirectory)))

	return dumper.execute(stringBuilder.String())
}
