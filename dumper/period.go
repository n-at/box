package dumper

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"sort"
)

type PeriodDump struct {
	name                string
	dumpType            Type
	rootPath            string
	fileName            string
	tmpDumpFileName     string
	tmpLogFileName      string
	tmpChecksumFileName string
	maxItemsCount       int
	overwrite           bool
}

func (period *PeriodDump) dumpFileName() string {
	return fmt.Sprintf("%s%c%s", period.rootPath, os.PathSeparator, period.fileName)
}

func (period *PeriodDump) logFileName() string {
	return fmt.Sprintf("%s%c%s.log", period.rootPath, os.PathSeparator, period.fileName)
}

func (period *PeriodDump) checksumFileName() string {
	return fmt.Sprintf("%s%c%s.checksum", period.rootPath, os.PathSeparator, period.fileName)
}

func (period *PeriodDump) exists() bool {
	_, err := os.Stat(period.dumpFileName())
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

func (period *PeriodDump) rotate() error {
	if period.maxItemsCount < 0 {
		return nil
	}

	files, err := os.ReadDir(period.rootPath)
	if err != nil {
		return err
	}

	logFileRegexp, err := regexp.Compile("^.+\\.log$")
	if err != nil {
		return err
	}
	checksumFileRegexp, err := regexp.Compile("^.+\\.checksum$")
	if err != nil {
		return err
	}

	var dumpFiles []string

	for _, file := range files {
		filename := file.Name()
		if logFileRegexp.MatchString(filename) || checksumFileRegexp.MatchString(filename) {
			continue
		}
		dumpFiles = append(dumpFiles, filename)
	}

	sort.Strings(dumpFiles)

	for i := 0; i < len(dumpFiles)-period.maxItemsCount; i++ {
		dumpFilePath := fmt.Sprintf("%s%c%s", period.rootPath, os.PathSeparator, dumpFiles[i])
		if err := os.Remove(dumpFilePath); err != nil {
			log.Errorf("%s (%s) %s: unable to delete dump file: %s", period.name, period.dumpType, period.fileName, err)
		}
		dumpChecksumPath := fmt.Sprintf("%s%c%s.checksum", period.rootPath, os.PathSeparator, dumpFiles[i])
		if err := os.Remove(dumpChecksumPath); err != nil {
			log.Errorf("%s (%s) %s: unable to delete checksum file: %s", period.name, period.dumpType, period.fileName, err)
		}
		dumpLogPath := fmt.Sprintf("%s%c%s.log", period.rootPath, os.PathSeparator, dumpFiles[i])
		if err := os.Remove(dumpLogPath); err != nil {
			log.Errorf("%s (%s) %s: unable to delete log file: %s", period.name, period.dumpType, period.fileName, err)
		}
	}

	return nil
}

func (period *PeriodDump) execute() error {
	if period.exists() && !period.overwrite {
		log.Infof("%s (%s) %s: already exists, skipping", period.name, period.dumpType, period.fileName)
		return nil
	}

	if err := makeDirectory(period.rootPath); err != nil {
		return err
	}
	if err := copyFile(period.tmpDumpFileName, period.dumpFileName()); err != nil {
		return err
	}
	if err := copyFile(period.tmpLogFileName, period.logFileName()); err != nil {
		return err
	}
	if err := copyFile(period.tmpChecksumFileName, period.checksumFileName()); err != nil {
		return err
	}

	log.Infof("%s (%s) %s: done", period.name, period.dumpType, period.fileName)

	return nil
}
