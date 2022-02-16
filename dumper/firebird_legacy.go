package dumper

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type FirebirdLegacyDumper struct {
	AbstractDumper
}

func NewFirebirdLegacy(global GlobalConfiguration, local Configuration) (*FirebirdLegacyDumper, error) {
	if len(global.GbakExecutable) == 0 {
		return nil, errors.New("gbak executable not defined")
	}

	dumper := FirebirdLegacyDumper{
		AbstractDumper{
			globalConfiguration: global,
			configuration:       local,
			time:                time.Now(),
		},
	}

	return &dumper, nil
}

func (dumper *FirebirdLegacyDumper) Dump() error {
	stringBuilder := strings.Builder{}

	//https://github.com/FirebirdSQL/firebird/releases/tag/R2_5_9
	//Example configuration
	//host: "localhost"
	//port: 3050
	//username: "SYSDBA"
	//password: "masterkey"
	//db: /sqlbase/database.fdb

	vars := dumper.configuration.Vars

	stringBuilder.WriteString(fmt.Sprintf("\"%s\" -VERIFY -BACKUP_DATABASE -GARBAGE_COLLECT ", esc(dumper.globalConfiguration.GbakExecutable)))

	db, ok := vars["db"]
	if !ok {
		return errors.New("database path not defined")
	}

	user, ok := vars["username"]
	if ok {
		stringBuilder.WriteString(fmt.Sprintf("-USER \"%s\" ", esc(user)))
	}

	password, ok := vars["password"]
	if ok {
		stringBuilder.WriteString(fmt.Sprintf("-PASSWORD \"%s\" ", esc(password)))
	}

	host, okHost := vars["host"]
	port, okPort := vars["port"]

	var source string

	if okHost {
		if okPort {
			source = fmt.Sprintf("\"%s/%s:%s\" ", esc(host), esc(port), esc(db))
		} else {
			source = fmt.Sprintf("\"%s:%s\" ", esc(host), esc(db))
		}
	} else {
		source = fmt.Sprintf("\"%s\" ", db)
	}

	stringBuilder.WriteString(fmt.Sprintf("%s \"%s\"", source, esc(dumper.tmpDumpFileName())))

	return dumper.execute(stringBuilder.String())
}
