package walk

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func checkCntl(path string, isDir bool, out chan<- *file, cntl <-chan error) error {
	if err := <-cntl; err == cntlClose {
		return cntlClose
	} else if isDir && err == SkipRecurse {
		return fs.SkipDir
	} else if err != nil {
		out <- &file{Path: path, Err: fmt.Errorf("Unexpected error: %w", err)}
		return err
	}
	return nil
}

func recurseFile(r io.Reader, path string, out chan<- *file, cntl <-chan error) error {
	if err := checkCntl(path, true, out, cntl); err != nil && err != fs.SkipDir {
		return err
	}

	pl := strings.ToLower(path)
	if strings.HasSuffix(pl, ".zip") {
		return recurseZip(r, path, out, cntl)
	} else if strings.HasSuffix(pl, ".tar") {
		return recurseTar(r, path, out, cntl)
	} else if strings.HasSuffix(pl, ".tgz") || strings.HasSuffix(pl, ".tar.gz") {
		gr, err := gzip.NewReader(r)
		if err != nil {
			out <- &file{Path: path, Err: fmt.Errorf("Unable to create gzip reader: %w", err)}
			return nil
		}
		return recurseTar(gr, path, out, cntl)
	} else if strings.HasSuffix(pl, ".tbz") || strings.HasSuffix(pl, ".tar.bz2") {
		return recurseTar(bzip2.NewReader(r), path, out, cntl)
	}

	return nil
}

func recurseFS(rfs fs.FS, rootPath string, out chan<- *file, cntl <-chan error) error {
	return fs.WalkDir(rfs, ".", func(path string, d fs.DirEntry, err error) error {
		fp := filepath.Join(rootPath, path)

		if err != nil {
			out <- &file{Path: fp, Err: fmt.Errorf("Unable to walk file: %w", err)}
			return checkCntl(fp, false, out, cntl)
		}

		fi, err := d.Info()
		if err != nil {
			out <- &file{Path: fp, Err: fmt.Errorf("Unable to get FileInfo: %w", err)}
			return checkCntl(fp, false, out, cntl)
		}

		//skip root directory since it's already been visited
		if fi.Name() == "." {
			return nil
		}

		//file
		if !fi.IsDir() {
			f, err := rfs.Open(path)
			if err != nil {
				out <- &file{Path: fp, Err: fmt.Errorf("Unable to open file: %w", err)}
				return checkCntl(fp, false, out, cntl)
			}

			out <- &file{FileInfo: fi, Reader: f, Path: fp}
			err = recurseFile(f, fp, out, cntl)
			f.Close()
			return err
		}

		//directory
		out <- &file{FileInfo: fi, Path: fp}
		return checkCntl(fp, true, out, cntl)
	})
}

func recurseRoot(path string, out chan<- *file, cntl <-chan error) {
	if err := checkCntl(path, true, out, cntl); err != nil {
		close(out)
		return
	}

	fi, err := os.Stat(path)
	if err != nil {
		out <- &file{Path: path, Err: fmt.Errorf("Unable to get FileInfo: %w", err)}
		close(out)
		<-cntl
		return
	}

	//file
	if !fi.IsDir() {
		f, err := os.Open(path)
		if err != nil {
			out <- &file{Path: path, Err: fmt.Errorf("Unable to open file: %w", err)}
			close(out)
			<-cntl
			return
		}

		out <- &file{FileInfo: fi, Reader: f, Path: path}
		recurseFile(f, path, out, cntl)

		close(out)
		<-cntl
		f.Close()
		return
	}

	//directory
	out <- &file{FileInfo: fi, Path: path}
	if err := checkCntl(path, true, out, cntl); err != nil {
		close(out)
		<-cntl
		return
	}

	dfs := os.DirFS(path)
	recurseFS(dfs, path, out, cntl)
	close(out)
	<-cntl
}
