package walk

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

func recurseTar(r io.Reader, rootPath string, out chan<- *file, cntl <-chan error) error {
	skipDirs := make(map[string]struct{})
	tr := tar.NewReader(r)

outer:
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			out <- &file{Path: rootPath, Err: fmt.Errorf("Unable to parse tgz: %w", err)}
			return checkCntl(rootPath, false, out, cntl)
		}

		fi := header.FileInfo()

		//skip skipped directories
		for dir := filepath.Dir(header.Name); dir != "."; dir = filepath.Dir(dir) {
			if _, ok := skipDirs[dir]; ok {
				if fi.IsDir() {
					skipDirs[filepath.Clean(header.Name)] = struct{}{}
				}
				continue outer
			}
		}

		fp := filepath.Join(rootPath, header.Name)

		//directory
		if fi.IsDir() {
			if _, ok := skipDirs[header.Name]; ok {
				continue
			}
			out <- &file{FileInfo: fi, Path: fp}
			if err = checkCntl(fp, true, out, cntl); err == fs.SkipDir {
				skipDirs[filepath.Clean(header.Name)] = struct{}{}
			} else if err != nil {
				return err
			}
			continue
		}

		//file
		out <- &file{FileInfo: fi, Reader: tr, Path: fp}

		if err = recurseFile(tr, fp, out, cntl); err != nil {
			return err
		}
	}
	return nil
}
