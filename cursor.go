//package walk recursively walks through filesystems (including archive files like zip and tgz). Filters can be used to control visited files and folders.
package walk

import (
	"errors"
	"io"
	"io/fs"
)

var cntlNext error
var cntlClose error = errors.New("Close")
var Skip = errors.New("Skip")
var SkipRecurse = errors.New("SkipRecurse")

type file struct {
	FileInfo fs.FileInfo
	Reader   io.Reader
	Path     string
	Err      error
}

type Cursor struct {
	cur         *file
	in          chan *file
	cntl        chan error
	filters     map[string]CursorFilterFunc
	filterOrder []string
}

//New returns a new Cursor for the given path (directory or file).
//The Cursor is not thread-safe, and should not be used on untrusted data
func New(path string) *Cursor {
	in := make(chan *file)
	cntl := make(chan error)
	go recurseRoot(path, in, cntl)
	return &Cursor{in: in, cntl: cntl, filters: make(map[string]CursorFilterFunc), filterOrder: make([]string, 0)}
}

//CursorFilterFunc receives the fs.FileInfo, path, and error before it is returned by Next.
//If the error returned is nil, then the Cursor returns the file as normal.
//If Skip is returned, the Cursor will move to the next file, including inside the current directory or archive file.
//If SkipRecurse is returned, and the cursor is at a directory or archive file,
//the directory or archive file will be returned, but won't be recursed into.
//If any other error is returned, Next will receive io.EOF and the Cursor will be closed
type CursorFilterFunc func(fi fs.FileInfo, path string, err error) error

//RegisterFilterFunc registers the given CursorFilterFunc for the Cursor. Filters are processed in the order they are registered
func (c *Cursor) RegisterFilterFunc(name string, fn CursorFilterFunc) {
	if _, ok := c.filters[name]; !ok {
		c.filterOrder = append(c.filterOrder, name)
	}
	c.filters[name] = fn
}

//UnregisterFilterFunc unregisters the given CursorFilterFunc
func (c *Cursor) UnregisterFilterFunc(name string) {
	for idx, n := range c.filterOrder {
		if n == name {
			c.filterOrder = append(c.filterOrder[:idx], c.filterOrder[idx+1:]...)
			delete(c.filters, name)
			return
		}
	}
}

func (c *Cursor) next(cmd error) (fi fs.FileInfo, path string, err error) {
	c.cntl <- cmd
	f, ok := <-c.in
	if !ok {
		return nil, "", io.EOF
	}
	for _, name := range c.filterOrder {
		err := c.filters[name](f.FileInfo, f.Path, f.Err)
		switch err {
		case nil:
			continue
		case Skip:
			return c.Next()
		case SkipRecurse:
			return c.next(SkipRecurse)
		default:
			c.Close()
			return c.next(err)
		}
	}
	c.cur = f
	return f.FileInfo, f.Path, f.Err
}

//Next moves the Cursor to the next file or directory and returns the fs.FileInfo, full path, and an error if one occurred.
//Next returns io.EOF after the last file or directory has been read and the Cursor is closed.
func (c *Cursor) Next() (fi fs.FileInfo, path string, err error) {
	return c.next(cntlNext)
}

//Read implements io.Reader. Read should not be called if the last call to Next returns a directory or an error. If Next returns io.EOF, Read should not be called again.
func (c *Cursor) Read(b []byte) (int, error) {
	return c.cur.Reader.Read(b)
}

//Close can be used to close the cursor prematurely. The cursor is closed implicitly when Next returns io.EOF
func (c *Cursor) Close() {
	c.cntl <- cntlClose
}
