package dumper

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	HashMD5    = "md5"
	HashSha256 = "sha256"
	HashSha1   = "sha1"
)

type Dumper interface {
	Dump() error
}

///////////////////////////////////////////////////////////////////////////////

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

	log.Infof("%s: staring", dumper.configuration.Name)

	if err := dumper.executeCommand(commandline); err != nil {
		return err
	}

	log.Infof("%s: execution done", dumper.configuration.Name)

	if err := dumper.calculateChecksums(); err != nil {
		return err
	}

	log.Infof("%s: checksums calculated", dumper.configuration.Name)

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

	if err := dumper.clearTmpFiles(); err != nil {
		return err
	}

	log.Infof("%s: done", dumper.configuration.Name)

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

	cmd := exec.Command("sh", "-c", commandline)
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

///////////////////////////////////////////////////////////////////////////////

type PeriodDump struct {
	name                string
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

	files, err := ioutil.ReadDir(period.rootPath)
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
			log.Errorf("%s %s: unable to delete dump file: %s", period.name, period.fileName, err)
		}
		dumpChecksumPath := fmt.Sprintf("%s%c%s.checksum", period.rootPath, os.PathSeparator, dumpFiles[i])
		if err := os.Remove(dumpChecksumPath); err != nil {
			log.Errorf("%s %s: unable to delete checksum file: %s", period.name, period.fileName, err)
		}
		dumpLogPath := fmt.Sprintf("%s%c%s.log", period.rootPath, os.PathSeparator, dumpFiles[i])
		if err := os.Remove(dumpLogPath); err != nil {
			log.Errorf("%s %s: unable to delete log file: %s", period.name, period.fileName, err)
		}
	}

	return nil
}

func (period *PeriodDump) execute() error {
	if period.exists() && !period.overwrite {
		log.Infof("%s %s: already exists, skipping", period.name, period.fileName)
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
	if err := period.rotate(); err != nil {
		return err
	}

	log.Infof("%s %s: done", period.name, period.fileName)

	return nil
}

///////////////////////////////////////////////////////////////////////////////

func makeDirectory(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return nil
}

func fileChecksum(hashType, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var hasher hash.Hash
	switch hashType {
	case HashMD5:
		hasher = md5.New()
	case HashSha1:
		hasher = sha1.New()
	case HashSha256:
		hasher = sha256.New()
	default:
		return "", errors.New("unknown hash type")
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func esc(param string) string {
	param = strings.ReplaceAll(param, "$", "\\$")
	param = strings.ReplaceAll(param, "\"", "\\\"")
	param = strings.ReplaceAll(param, "\n", "\\n")
	return param
}
