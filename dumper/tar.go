package dumper

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type TarDumper struct {
	AbstractDumper
}

func NewTar(global GlobalConfiguration, local Configuration) (*TarDumper, error) {
	if len(global.TarExecutable) == 0 {
		return nil, errors.New("tar executable not defined")
	}

	dumper := TarDumper{
		AbstractDumper{
			globalConfiguration: global,
			configuration:       local,
			time:                time.Now(),
		},
	}

	return &dumper, nil
}

func (d *TarDumper) Dump() error {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("\"%s\" ", esc(d.globalConfiguration.TarExecutable)))
	sb.WriteString("--verbose ")
	sb.WriteString("--create ")

	//Example configuration:
	//path: "/directory/location"
	//compress: "none|bzip2|gzip|lzma|xz"

	vars := d.configuration.Vars

	path, ok := vars["path"]
	if !ok || len(path) == 0 {
		return errors.New("path not defined")
	}

	compress, ok := vars["compress"]
	if !ok || len(compress) == 0 {
		compress = "none"
	}

	switch compress {
	case "none":
		sb.WriteString("")
		break
	case "bzip2":
		sb.WriteString("--bzip2 ")
		break
	case "gzip":
		sb.WriteString("--gzip ")
		break
	case "lzma":
		sb.WriteString("--lzma ")
		break
	case "xz":
		sb.WriteString("--xz ")
		break
	}

	directory, targetFile := splitTargetPath(path)
	if len(targetFile) == 0 {
		return errors.New("empty path target name")
	}

	sb.WriteString(fmt.Sprintf("--file \"%s\" ", esc(d.tmpDumpFileName())))

	if len(directory) > 0 {
		sb.WriteString(fmt.Sprintf("--directory \"%s\" ", esc(directory)))
	}

	for key, value := range vars {
		if key == "path" || key == "compress" || key == "verbose" || key == "create" || key == "directory" {
			continue
		}
		sb.WriteString(formatParam(key, value))
		sb.WriteString(" ")
	}

	sb.WriteString(fmt.Sprintf("\"%s\"", esc(targetFile)))

	return d.execute(sb.String())
}

func splitTargetPath(path string) (string, string) {
	sep := string(os.PathSeparator)
	parts := strings.Split(path, sep)

	for len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}

	s := len(parts)

	target := parts[s-1]
	directory := strings.Join(parts[:s-1], sep)

	if strings.HasPrefix(path, sep) && len(directory) == 0 {
		directory = sep
	}

	return directory, target
}
