package dumper

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

type Dumper interface {
	Dump() error
}

type AbstractDumper struct {
	globalConfiguration GlobalConfiguration
	configuration       Configuration
	time                time.Time
}

func (dumper *AbstractDumper) execute(commandline string) error {
	if len(dumper.configuration.Name) == 0 {
		return errors.New("dumper name not defined")
	}
	if len(dumper.configuration.Path) == 0 && len(dumper.globalConfiguration.Path) == 0 {
		return errors.New("dumper path not defined")
	}
	if len(dumper.configuration.TmpPath) == 0 && len(dumper.globalConfiguration.TmpPath) == 0 {
		return errors.New("dumper tmp path not defined")
	}

	log.Infof("%s (%s) starting...", dumper.configuration.Name, dumper.configuration.Type)

	if err := dumper.executeCommand(commandline); err != nil {
		return err
	}

	log.Infof("%s (%s) execution done", dumper.configuration.Name, dumper.configuration.Type)

	if err := dumper.calculateChecksums(); err != nil {
		return err
	}

	log.Infof("%s (%s) checksums calculated", dumper.configuration.Name, dumper.configuration.Type)

	log.Infof("%s (%s) copy latest dump...", dumper.configuration.Name, dumper.configuration.Type)
	latest := PeriodDump{
		name:                dumper.configuration.Name,
		rootPath:            dumper.rootPath(),
		fileName:            "latest",
		tmpDumpFileName:     dumper.tmpDumpFileName(),
		tmpLogFileName:      dumper.tmpLogFileName(),
		tmpChecksumFileName: dumper.tmpChecksumFileName(),
		maxItemsCount:       -1,
		overwrite:           true,
	}
	if err := latest.execute(); err != nil {
		return err
	}

	if dumper.configuration.Daily {
		log.Infof("%s (%s) copy daily dump...", dumper.configuration.Name, dumper.configuration.Type)
		daily := PeriodDump{
			name:                dumper.configuration.Name,
			rootPath:            fmt.Sprintf("%s%c%s", dumper.rootPath(), os.PathSeparator, "daily"),
			fileName:            dumper.time.Format("2006-01-02"),
			tmpDumpFileName:     dumper.tmpDumpFileName(),
			tmpLogFileName:      dumper.tmpLogFileName(),
			tmpChecksumFileName: dumper.tmpChecksumFileName(),
			maxItemsCount:       dumper.configuration.Days,
			overwrite:           false,
		}
		if err := daily.execute(); err != nil {
			return err
		}
	}

	if dumper.configuration.Weekly {
		log.Infof("%s (%s) copy weekly dump...", dumper.configuration.Name, dumper.configuration.Type)
		year, week := dumper.time.ISOWeek()
		weekly := PeriodDump{
			name:                dumper.configuration.Name,
			rootPath:            fmt.Sprintf("%s%c%s", dumper.rootPath(), os.PathSeparator, "weekly"),
			fileName:            fmt.Sprintf("%04d-%02d", year, week),
			tmpDumpFileName:     dumper.tmpDumpFileName(),
			tmpLogFileName:      dumper.tmpLogFileName(),
			tmpChecksumFileName: dumper.tmpChecksumFileName(),
			maxItemsCount:       dumper.configuration.Weeks,
			overwrite:           false,
		}
		if err := weekly.execute(); err != nil {
			return err
		}
	}

	if dumper.configuration.Monthly {
		log.Infof("%s (%s) copy monthly dump...", dumper.configuration.Name, dumper.configuration.Type)
		monthly := PeriodDump{
			rootPath:            fmt.Sprintf("%s%c%s", dumper.rootPath(), os.PathSeparator, "monthly"),
			fileName:            dumper.time.Format("2006-01"),
			tmpDumpFileName:     dumper.tmpDumpFileName(),
			tmpLogFileName:      dumper.tmpLogFileName(),
			tmpChecksumFileName: dumper.tmpChecksumFileName(),
			maxItemsCount:       dumper.configuration.Months,
			overwrite:           false,
		}
		if err := monthly.execute(); err != nil {
			return err
		}
	}

	log.Infof("%s (%s) clear tmp files...", dumper.configuration.Name, dumper.configuration.Type)
	if err := dumper.clearTmpFiles(); err != nil {
		return err
	}

	log.Infof("%s (%s) done", dumper.configuration.Name, dumper.configuration.Type)

	return nil
}

func (dumper *AbstractDumper) rootPath() string {
	if len(dumper.configuration.Path) != 0 {
		return dumper.configuration.Path
	}
	if len(dumper.globalConfiguration.Path) != 0 {
		return fmt.Sprintf("%s%c%s", dumper.globalConfiguration.Path, os.PathSeparator, dumper.configuration.Name)
	}
	return ""
}

func (dumper *AbstractDumper) tmpPath() string {
	if len(dumper.configuration.TmpPath) != 0 {
		return dumper.configuration.TmpPath
	}
	if len(dumper.globalConfiguration.TmpPath) != 0 {
		return dumper.globalConfiguration.TmpPath
	}
	return ""
}

func (dumper *AbstractDumper) tmpDumpFileName() string {
	return fmt.Sprintf("%s%c%s", dumper.tmpPath(), os.PathSeparator, dumper.configuration.Name)
}

func (dumper *AbstractDumper) tmpLogFileName() string {
	return fmt.Sprintf("%s%c%s.log", dumper.tmpPath(), os.PathSeparator, dumper.configuration.Name)
}

func (dumper *AbstractDumper) tmpChecksumFileName() string {
	return fmt.Sprintf("%s%c%s.checksum", dumper.tmpPath(), os.PathSeparator, dumper.configuration.Name)
}

func (dumper *AbstractDumper) executeCommand(commandline string) error {
	logFile, err := os.OpenFile(dumper.tmpLogFileName(), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(dumper.globalConfiguration.ShExecutable, "-c", commandline)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		return err
	}

	stat, err := os.Stat(dumper.tmpDumpFileName())
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return errors.New("empty dump file")
	} else {
		log.Infof("%s (%s) dump file size: %s", dumper.configuration.Name, dumper.configuration.Type, formatFileSize(stat.Size()))
	}

	return nil
}

func (dumper *AbstractDumper) calculateChecksums() error {
	md5Hash, err := fileChecksum(HashMD5, dumper.tmpDumpFileName())
	if err != nil {
		return err
	}

	sha1Hash, err := fileChecksum(HashSha1, dumper.tmpDumpFileName())
	if err != nil {
		return err
	}

	sha256Hash, err := fileChecksum(HashSha256, dumper.tmpDumpFileName())
	if err != nil {
		return err
	}

	output := fmt.Sprintf("MD5: %s\nSHA1: %s\nSHA256: %s\n", md5Hash, sha1Hash, sha256Hash)

	if err := os.WriteFile(dumper.tmpChecksumFileName(), []byte(output), 0644); err != nil {
		return err
	}

	return nil
}

func (dumper *AbstractDumper) clearTmpFiles() error {
	if err := os.Remove(dumper.tmpDumpFileName()); err != nil {
		return err
	}
	if err := os.Remove(dumper.tmpLogFileName()); err != nil {
		return err
	}
	if err := os.Remove(dumper.tmpChecksumFileName()); err != nil {
		return err
	}
	return nil
}
