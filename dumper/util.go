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
	"strings"
)

const (
	HashMD5    = "md5"
	HashSha256 = "sha256"
	HashSha1   = "sha1"
)

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

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	floatSize := float64(size) / 1024.0
	if floatSize < 1024 {
		return fmt.Sprintf("%.2f kB", floatSize)
	}
	floatSize /= 1024.0
	if floatSize < 1024 {
		return fmt.Sprintf("%.2f MB", floatSize)
	}
	return fmt.Sprintf("%.2f GB", floatSize/1024.0)
}
