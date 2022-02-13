package dumper

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
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

type AbstractDumper struct {
	globalConfiguration GlobalConfiguration
	configuration       Configuration
	time                time.Time
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

func (dumper *AbstractDumper) tmpPath() string {
	if len(dumper.configuration.TmpPath) != 0 {
		return dumper.configuration.TmpPath
	}
	if len(dumper.globalConfiguration.TmpPath) != 0 {
		return dumper.globalConfiguration.TmpPath
	}
	return ""
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) latestDumpFileName() string {
	return fmt.Sprintf("%s%clatest", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) latestLogFileName() string {
	return fmt.Sprintf("%s%clatest.log", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) latestChecksumFileName() string {
	return fmt.Sprintf("%s%clatest.checksum", dumper.rootPath(), os.PathSeparator)
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) dailyPath() string {
	return fmt.Sprintf("%s%cdaily", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) dailyName() string {
	return dumper.time.Format("2006-01-02")
}

func (dumper *AbstractDumper) dailyDumpFileName() string {
	return fmt.Sprintf("%s%c%s", dumper.dailyPath(), os.PathSeparator, dumper.dailyName())
}

func (dumper *AbstractDumper) dailyLogFileName() string {
	return fmt.Sprintf("%s%c%s.log", dumper.dailyPath(), os.PathSeparator, dumper.dailyName())
}

func (dumper *AbstractDumper) dailyChecksumFileName() string {
	return fmt.Sprintf("%s%c%s.checksum", dumper.dailyPath(), os.PathSeparator, dumper.dailyName())
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) weeklyPath() string {
	return fmt.Sprintf("%s%cweekly", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) weeklyName() string {
	year, week := dumper.time.ISOWeek()
	return fmt.Sprintf("%04d-%02d", year, week)
}

func (dumper *AbstractDumper) weeklyDumpFileName() string {
	return fmt.Sprintf("%s%c%s", dumper.weeklyPath(), os.PathSeparator, dumper.weeklyName())
}

func (dumper *AbstractDumper) weeklyLogFileName() string {
	return fmt.Sprintf("%s%c%s.log", dumper.weeklyPath(), os.PathSeparator, dumper.weeklyName())
}

func (dumper *AbstractDumper) weeklyChecksumName() string {
	return fmt.Sprintf("%s%c%s.checksum", dumper.weeklyPath(), os.PathSeparator, dumper.weeklyName())
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) monthlyPath() string {
	return fmt.Sprintf("%s%cmonthly", dumper.rootPath(), os.PathSeparator)
}

func (dumper *AbstractDumper) monthlyName() string {
	return dumper.time.Format("2006-01")
}

func (dumper *AbstractDumper) monthlyDumpFileName() string {
	return fmt.Sprintf("%s%c%s", dumper.monthlyPath(), os.PathSeparator, dumper.monthlyName())
}

func (dumper *AbstractDumper) monthlyLogFileName() string {
	return fmt.Sprintf("%s%c%s.log", dumper.monthlyPath(), os.PathSeparator, dumper.monthlyName())
}

func (dumper *AbstractDumper) monthlyChecksumName() string {
	return fmt.Sprintf("%s%c%s.checksum", dumper.monthlyPath(), os.PathSeparator, dumper.monthlyName())
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) tmpDumpFileName() string {
	return fmt.Sprintf("%s%c%s", dumper.tmpPath(), os.PathSeparator, dumper.configuration.Name)
}

func (dumper *AbstractDumper) tmpLogFileName() string {
	return fmt.Sprintf("%s%c%s.log", dumper.tmpPath(), os.PathSeparator, dumper.configuration.Name)
}

func (dumper *AbstractDumper) tmpChecksumFileName() string {
	return fmt.Sprintf("%s%c%s.checksum", dumper.tmpPath(), os.PathSeparator, dumper.configuration.Name)
}

///////////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) calculateChecksums() error {
	md5Hash, err := dumper.dumpHash(HashMD5)
	if err != nil {
		return err
	}

	sha1Hash, err := dumper.dumpHash(HashSha1)
	if err != nil {
		return err
	}

	sha256Hash, err := dumper.dumpHash(HashSha256)
	if err != nil {
		return err
	}

	output := fmt.Sprintf("MD5: %s\nSHA1: %s\nSHA256: %s\n", md5Hash, sha1Hash, sha256Hash)

	if err := os.WriteFile(dumper.tmpChecksumFileName(), []byte(output), 0644); err != nil {
		return err
	}

	return nil
}

func (dumper *AbstractDumper) dumpHash(hashType string) (string, error) {
	file, err := os.Open(dumper.tmpDumpFileName())
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

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) copyLatest() error {
	if err := copyFile(dumper.tmpDumpFileName(), dumper.latestDumpFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpLogFileName(), dumper.latestLogFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpChecksumFileName(), dumper.latestChecksumFileName()); err != nil {
		return err
	}
	return nil
}

func (dumper *AbstractDumper) copyDaily() error {
	if err := copyFile(dumper.tmpDumpFileName(), dumper.dailyDumpFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpLogFileName(), dumper.dailyLogFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpChecksumFileName(), dumper.dailyChecksumFileName()); err != nil {
		return err
	}
	return nil
}

func (dumper *AbstractDumper) copyWeekly() error {
	if err := copyFile(dumper.tmpDumpFileName(), dumper.weeklyDumpFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpLogFileName(), dumper.weeklyLogFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpChecksumFileName(), dumper.weeklyChecksumName()); err != nil {
		return err
	}
	return nil
}

func (dumper *AbstractDumper) copyMonthly() error {
	if err := copyFile(dumper.tmpDumpFileName(), dumper.monthlyDumpFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpLogFileName(), dumper.monthlyLogFileName()); err != nil {
		return err
	}
	if err := copyFile(dumper.tmpChecksumFileName(), dumper.monthlyChecksumName()); err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (dumper *AbstractDumper) rotateDaily() error {
	return nil //TODO
}

func (dumper *AbstractDumper) rotateWeekly() error {
	return nil //TODO
}

func (dumper *AbstractDumper) rotateMonthly() error {
	return nil //TODO
}

///////////////////////////////////////////////////////////////////////////////

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
