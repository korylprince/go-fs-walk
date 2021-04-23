package walk

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

func recurseZip(r io.Reader, path string, out chan<- *file, cntl <-chan error) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		out <- &file{Path: path, Err: fmt.Errorf("Unable to buffer zip file: %w", err)}
		return nil
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		out <- &file{Path: path, Err: fmt.Errorf("Unable to create zip file reader: %w", err)}
		return nil
	}

	return recurseFS(zr, path, out, cntl)
}
