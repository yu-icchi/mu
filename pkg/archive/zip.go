package archive

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var errIllegalFilePath = errors.New("illegal file path")

type zipArchiver struct{}

func NewZipArchiver() Archive {
	return &zipArchiver{}
}

func (z *zipArchiver) Decompress(dest, src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = r.Close()
	}()

	for _, file := range r.File {
		if err := z.unzip(dest, file); err != nil {
			return err
		}
	}
	return nil
}

func (z *zipArchiver) unzip(dest string, src *zip.File) error {
	r, err := src.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = r.Close()
	}()

	path := filepath.Join(dest, src.Name) // #nosec G305
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return errIllegalFilePath
	}

	if src.FileInfo().IsDir() {
		if err := os.MkdirAll(path, src.Mode()); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), src.Mode()); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, src.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = io.Copy(file, r) // #nosec G110
	return err
}
