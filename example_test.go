package walk_test

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	walk "github.com/korylprince/go-fs-walk"
)

func UnluckyFilter(fi fs.FileInfo, path string, err error) error {
	//skip files and folders named 13
	if strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name())) == "13" {
		return walk.SkipRecurse
	}
	return nil
}

func Example_cursorWalk() {
	c := walk.New("/path/to/root")
	c.RegisterFilterFunc("unlucky", UnluckyFilter)
	for fi, path, err := c.Next(); err != io.EOF; fi, path, err = c.Next() {
		if err != nil {
			panic(err)
		}
		fmt.Printf("Found lucky file %s at %s", fi.Name(), path)
		//use c as io.Reader...
	}
}
