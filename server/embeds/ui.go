package embeds

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:generate cp -r ../../client/dashboard/build dashboard
//go:generate tar -zcf eastnor.tgz -C ../../themes/eastnor .
//go:embed dashboard/**
var dashboardFS embed.FS

//go:embed eastnor.tgz
var themeFS embed.FS

func DashboardFS(path string) (fs.File, error) {

	return dashboardFS.Open("dashboard/" + path)

}

func CopyTheme(templatePath string) {
	os.MkdirAll(templatePath, os.ModePerm)
	f, _ := themeFS.Open("eastnor.tgz")
	Untar(templatePath, f)
}

func Untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
