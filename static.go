package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type localFS struct{}

var local localFS

type staticFS struct{}

var static staticFS

type file struct {
	compressed string
	size       int64
	local      string

	data []byte
	once sync.Once
	name string
}

func (_ localFS) Open(name string) (http.File, error) {
	f, present := data[name]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_ staticFS) Open(name string) (http.File, error) {
	f, present := data[name]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		var gr *gzip.Reader
		gr, err = gzip.NewReader(bytes.NewBufferString(f.compressed))
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
		f.name = path.Base(name)
	})
	return f.File()
}

func (f *file) File() (http.File, error) {
	hf := httpFile{
		Reader: bytes.NewReader(f.data),
		file:   f,
	}
	return &hf, nil
}

type httpFile struct {
	*bytes.Reader
	*file
}

func (f *file) Close() error {
	return nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *file) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return f.size
}

func (f *file) Mode() os.FileMode {
	return 0
}

func (f *file) ModTime() time.Time {
	return time.Time{}
}
func (f *file) IsDir() bool {
	return false
}
func (f *file) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If local is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return local
	}
	return static
}

var data = map[string]*file{}
