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

	latest  PeriodDump
	daily   PeriodDump
	weekly  PeriodDump
	monthly PeriodDump
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

	dumper.latest = PeriodDump{
		name:                dumper.configuration.Name,
		dumpType:            dumper.configuration.Type,
		rootPath:            dumper.rootPath(),
		fileName:            "latest",
		tmpDumpFileName:     dumper.tmpDumpFileName(),
		tmpLogFileName:      dumper.tmpLogFileName(),
		tmpChecksumFileName: dumper.tmpChecksumFileName(),
		maxItemsCount:       -1,
		overwrite:           true,
	}
	dumper.daily = PeriodDump{
		name:                dumper.configuration.Name,
		dumpType:            dumper.configuration.Type,
		rootPath:            fmt.Sprintf("%s%c%s", dumper.rootPath(), os.PathSeparator, "daily"),
		fileName:            dumper.dailyFileName(),
		tmpDumpFileName:     dumper.tmpDumpFileName(),
		tmpLogFileName:      dumper.tmpLogFileName(),
		tmpChecksumFileName: dumper.tmpChecksumFileName(),
		maxItemsCount:       dumper.configuration.Days,
		overwrite:           false,
	}
	dumper.weekly = PeriodDump{
		name:                dumper.configuration.Name,
		dumpType:            dumper.configuration.Type,
		rootPath:            fmt.Sprintf("%s%c%s", dumper.rootPath(), os.PathSeparator, "weekly"),
		fileName:            dumper.weeklyFileName(),
		tmpDumpFileName:     dumper.tmpDumpFileName(),
		tmpLogFileName:      dumper.tmpLogFileName(),
		tmpChecksumFileName: dumper.tmpChecksumFileName(),
		maxItemsCount:       dumper.configuration.Weeks,
		overwrite:           false,
	}
	dumper.monthly = PeriodDump{
		name:                dumper.configuration.Name,
		dumpType:            dumper.configuration.Type,
		rootPath:            fmt.Sprintf("%s%c%s", dumper.rootPath(), os.PathSeparator, "monthly"),
		fileName:            dumper.monthlyFileName(),
		tmpDumpFileName:     dumper.tmpDumpFileName(),
		tmpLogFileName:      dumper.tmpLogFileName(),
		tmpChecksumFileName: dumper.tmpChecksumFileName(),
		maxItemsCount:       dumper.configuration.Months,
		overwrite:           false,
	}

	dumpNeeded := dumper.isDumpNeeded()

	if dumpNeeded {
		defer func() {
			log.Infof("%s (%s) clear tmp files...", dumper.configuration.Name, dumper.configuration.Type)
			if err := dumper.clearTmpFiles(); err != nil {
				log.Errorf("%s (%s) clear tmp files error: %s", dumper.configuration.Name, dumper.configuration.Type, err)
			}
		}()

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
		if err := dumper.latest.execute(); err != nil {
			return err
		}
	} else {
		log.Infof("%s (%s) no dump needed, skipping", dumper.configuration.Name, dumper.configuration.Type)
	}

	if dumper.configuration.Daily {
		if dumpNeeded {
			log.Infof("%s (%s) copy daily dump...", dumper.configuration.Name, dumper.configuration.Type)
			if err := dumper.daily.execute(); err != nil {
				return err
			}
		}
		if err := dumper.daily.rotate(); err != nil {
			return err
		}
	}

	if dumper.configuration.Weekly {
		if dumpNeeded {
			log.Infof("%s (%s) copy weekly dump...", dumper.configuration.Name, dumper.configuration.Type)
			if err := dumper.weekly.execute(); err != nil {
				return err
			}
		}
		if err := dumper.weekly.rotate(); err != nil {
			return err
		}
	}

	if dumper.configuration.Monthly {
		if dumpNeeded {
			log.Infof("%s (%s) copy monthly dump...", dumper.configuration.Name, dumper.configuration.Type)
			if err := dumper.monthly.execute(); err != nil {
				return err
			}
		}
		if err := dumper.monthly.rotate(); err != nil {
			return err
		}
	}

	log.Infof("%s (%s) done", dumper.configuration.Name, dumper.configuration.Type)

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) isDumpNeeded() bool {
	if dumper.configuration.ForceLatest {
		return true
	}
	if dumper.configuration.Daily && !dumper.daily.exists() {
		return true
	}
	if dumper.configuration.Weekly && !dumper.weekly.exists() {
		return true
	}
	if dumper.configuration.Monthly && !dumper.monthly.exists() {
		return true
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////

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

func (dumper *AbstractDumper) dailyFileName() string {
	return dumper.time.Format("2006-01-02")
}

func (dumper *AbstractDumper) weeklyFileName() string {
	year, week := dumper.time.ISOWeek()
	return fmt.Sprintf("%04d-%02d", year, week)
}

func (dumper *AbstractDumper) monthlyFileName() string {
	return dumper.time.Format("2006-01")
}

///////////////////////////////////////////////////////////////////////////////

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
