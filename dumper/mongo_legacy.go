package dumper

import (
	"errors"
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
	//https://docs.mongodb.com/v4.0/reference/program/mongodump/
	//Compatible with MongoDB 4.0-2.6
	//Example configuration:
	//host: "localhost"
	//port: "27017"
	//username: "admin"
	//password: "******"
	//authenticationDatabase: "admin"
	//db: "users"

	commandline := buildMongoCommandline(dumper.globalConfiguration.Mongodump4Executable, dumper.tmpDumpFileName(), dumper.configuration.Vars)

	return dumper.execute(commandline)
}
