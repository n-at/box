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
	//databases: "database"
	vars := d.configuration.Vars

	for key, value := range vars {
		if key == "verbose" || key == "help" {
			continue
		}
		sb.WriteString(fmt.Sprintf("--%s=\"%s\" ", key, esc(value)))
	}

	sb.WriteString(fmt.Sprintf(" | gzip > \"%s\"", esc(d.tmpDumpFileName())))

	return d.execute(sb.String())
}
