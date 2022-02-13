package dumper

import (
	"errors"
	"time"
)

type PostgresDumper struct {
	AbstractDumper
}

func NewPostgres(global GlobalConfiguration, local Configuration) (*PostgresDumper, error) {
	if len(global.PgdumpExecutable) == 0 {
		return nil, errors.New("pg_dump executable not defined")
	}
	if len(local.Name) == 0 {
		return nil, errors.New("dumper name not defined")
	}
	if len(global.Path) == 0 && len(local.Path) == 0 {
		return nil, errors.New("dumper path not defined")
	}
	if len(global.TmpPath) == 0 && len(local.TmpPath) == 0 {
		return nil, errors.New("dumper tmp path new defined")
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
	if err := dumper.makeDirectories(); err != nil {
		return err
	}

	//TODO make command string
	//TODO execute command
	//TODO copy dump
	//TODO rotate dumps

	return nil
}
