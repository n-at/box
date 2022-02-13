package dumper

import (
	"errors"
	"fmt"
	"os"
)

type Dumper interface {
	Dump() error
}

type AbstractDumper struct {
	globalConfiguration GlobalConfiguration
	configuration       Configuration
}

func (dumper *AbstractDumper) rootPath() string {
	if len(dumper.configuration.Path) != 0 {
		return dumper.configuration.Path
	}
	if len(dumper.globalConfiguration.Path) != 0 {
		return dumper.globalConfiguration.Path
	}
	return ""
}

func (dumper *AbstractDumper) dailyPath() string {
	return fmt.Sprintf("%s%cdaily", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) weeklyPath() string {
	return fmt.Sprintf("%s%cweekly", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) monthlyPath() string {
	return fmt.Sprintf("%s%cmonthly", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) makeDirectories() error {
	if err := makeDirectory(dumper.rootPath()); err != nil {
		return err
	}
	if dumper.configuration.Daily {
		if err := makeDirectory(dumper.dailyPath()); err != nil {
			return err
		}
	}
	if dumper.configuration.Weekly {
		if err := makeDirectory(dumper.weeklyPath()); err != nil {
			return err
		}
	}
	if dumper.configuration.Monthly {
		if err := makeDirectory(dumper.monthlyPath()); err != nil {
			return err
		}
	}
	return nil
}

func makeDirectory(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}
