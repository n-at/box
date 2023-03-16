package dumper

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type MysqlDumper struct {
	AbstractDumper
}

func NewMysql(global GlobalConfiguration, local Configuration) (*MysqlDumper, error) {
	if len(global.MysqldumpExecutable) == 0 {
		return nil, errors.New("mysqldump executable not defined")
	}

	dumper := MysqlDumper{
		AbstractDumper{
			globalConfiguration: global,
			configuration:       local,
			time:                time.Now(),
		},
	}

	return &dumper, nil
}

func (d *MysqlDumper) Dump() error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("\"%s\" --verbose ", d.globalConfiguration.MysqldumpExecutable))

	//https://mariadb.com/kb/en/mariadb-dumpmysqldump/
	//Example configuration:
	//host: "localhost"
	//port: "3306"
	//user: "user"
	//password: "******"
	//database: "database"
	vars := d.configuration.Vars

	database, ok := vars["database"]
	if !ok || len(database) == 0 {
		return errors.New("database name required")
	}

	for key, value := range vars {
		if key == "verbose" || key == "help" || key == "databases" || key == "all-databases" || key == "database" {
			continue
		}
		sb.WriteString(formatParam(key, value))
		sb.WriteString(" ")
	}

	sb.WriteString(fmt.Sprintf(" \"%s\" | gzip > \"%s\"", esc(database), esc(d.tmpDumpFileName())))

	return d.execute(sb.String())
}
