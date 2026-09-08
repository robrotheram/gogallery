package config

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	if info.Size() == 0 {
		return false
	}
	return !info.IsDir()
}

func Dir(src string, dst string) error {
	var err error
	var fds []fs.DirEntry
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = os.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := filepath.Join(src, fd.Name())
		dstfp := filepath.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = Dir(srcfp, dstfp); err != nil {
				return fmt.Errorf("copy directory %q: %w", srcfp, err)
			}
		} else {
			if err = Copy(srcfp, dstfp); err != nil {
				return fmt.Errorf("copy file %q: %w", srcfp, err)
			}
		}
	}
	return nil
}

func Copy(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil { // #nosec G304 -- source is a user-configured local gallery file
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil { // #nosec G304 -- destination is under the validated build directory
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
