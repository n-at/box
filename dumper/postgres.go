package dumper

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type PostgresDumper struct {
	AbstractDumper
}

func NewPostgres(global GlobalConfiguration, local Configuration) (*PostgresDumper, error) {
	if len(global.PgdumpExecutable) == 0 {
		return nil, errors.New("pg_dump executable not defined")
	}

	dumper := PostgresDumper{
		AbstractDumper{
			globalConfiguration: global,
			configuration:       local,
			time:                time.Now(),
		},
	}

	return &dumper, nil
}

func (dumper *PostgresDumper) Dump() error {
	sb := strings.Builder{}

	//https://www.postgresql.org/docs/14/app-pgdump.html
	//Example configuration:
	//password: "******"
	//schema: "public"
	//dbname: "database"
	//host: "localhost"
	//port: "5432"
	//username: "user"
	vars := dumper.configuration.Vars

	if _, ok := vars["password"]; ok {
		password := fmt.Sprintf("PGPASSWORD=\"%s\" ", esc(vars["password"]))
		sb.WriteString(password)
	}

	sb.WriteString(fmt.Sprintf("\"%s\" ", esc(dumper.globalConfiguration.PgdumpExecutable)))
	sb.WriteString("--verbose ")
	sb.WriteString("--format=plain ")

	for key, value := range vars {
		if key == "verbose" || key == "format" || key == "password" {
			continue
		}
		sb.WriteString(formatParam(key, value))
		sb.WriteString(" ")
	}

	sb.WriteString(fmt.Sprintf("| gzip > \"%s\"", esc(dumper.tmpDumpFileName())))

	return dumper.execute(sb.String())
}
