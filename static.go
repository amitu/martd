package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDir struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDir) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDir{fs: _escLocal, name: name}
	}
	return _escDir{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/client.js": {
		local:   "client.js",
		size:    2698,
		modtime: 1446020142,
		compressed: `
H4sIAAAJbogA/4xWbW/bNhD+LP8Kzh8qapHlOO5SwEFWDEWwbliaIvaGAEFQ0DJty5FJj6RiG23+++4o
ipb8EiwfYpG8l+cePndSt0vmxqz0oNudZdoks8zMi3GSymX3gUuRKql198PlZf/iQ7/VWmdiItfJkikz
IddkWojUZFLQiHxvBS9MEbZgm9oBoYXKYzJhhsUkZXk+ZulzTLhhM+tifdBB8DU180wnD7d/fQY89/zf
gmtDfvwgv0GkF/5wN17w1EQ0vB2CzYU1HI2+Jv3kPIyuINQmkSsuKOYiH0n49W44CsmAhL/fjMKYWBw9
a5hNCd0hAD/Njcv3mbMJVzT8Y9r5IgXv3DKTzkMHGH1fW8cdPklhuDAds11xsA/ZapVnKUMOupvOer3u
TKVadgAFF6mc8InHLBSE2GrDDE/nTMx4gz6P0VoN0Yr8Svrk3TtPZ/2ZoqFeSaH5iG9MTDYl6isHW0ws
P3ZTcVMoQTbwDOetwCfV711aZwEkzJNpLqWiiIX2yFm5pxioYUmjiPxMzje9c/gDgygxcmhUJma0dxkl
uhhrt8K0r61SKLMiO1RQldHmsTDOiPsJO+Fbi5pD47D2g9CqWhFBJeLvyI5dJMOb+39u7mGvjT0BLZFL
YHYutRn88r5/0Wt7S+X0Ccot8txv4w0Knutm2NSWihXTyG/yF66+jYvlilseWK65PwOtpMDYtzSXmu9O
q2OgtE4dJoXuGtf7CjX+U7nE/2B+jv62cAcS9pqgH/HhqXJ2u06AOx+7DARb8oGNFNs1JhnYVOW6EqQe
ABG4Y/8dywch3UbZXw7krMlZBSDxgR/BwjqPbWGI+YBYB3535q7N7TfvMmFjqUyZLThxEUYV/Koq55VA
EbwMdexOvTHu0MiX50Ru/aCIAdZqWUuZSHk+2G+KIJjwnEPjH+fAw9k1FyY8bC48gQGE8u6ChD6CLK/b
0BxepIgQKiZWUSQTe/qosVkO4+SZbzU9dqk7iFECE29m5jC1zqtyEMTZNaHtd5jeJjsj7TqWZjA/fas6
D5oQXzu00cNn5cS3nRPXBqqBsRjV7+ygk4PAqK1DipThQIXTP4d3X5IVU5qXMUqZNOhCy322XDcdq8tb
BCkMbJnzJJcz2v5bPAu5FtVtD0g7ts/RzthkwknLNZYDkh2gcAyu2DaXbLLLiIW5TajtDZfH7MklKlMs
DnVxcOU+S43J4ET718S8eKIVziol9FiKr2B4YStVi9tgzG8G7U/VO/FGKaksdeAYV6WWRMZk4V1qmVqN
X/dzUo8nWMOjMmQZwH0wwQfDKFtyWRiKvRlDL5R6PlLeblo1po+v/s1XRPA/kjYGV5mv1hX7ctxjsqb9
iqNT6fCDIKqRYZu3/AaoZqCt5ar1itPxvwAAAP//qxOKIYoKAAA=
`,
	},

	"/index.html": {
		local:   "index.html",
		size:    528,
		modtime: 1446019182,
		compressed: `
H4sIAAAJbogA/3ySv2rDMBCHZ+cpDtHBocFyPSaKl75CobMiqbGKLBnp3FBK3736Y4ozONMd5+/8++DE
BhxNv6vYoLiMtWKo0ah+5B4lSDU6RsskMnSB2MXJ7wwH4fWEELw4E2G0sth8BtIzWj6smNRWlMJTTdI2
2Td8mpSVNWGTV/0716jtNfds8D3Zn3Zp44t7GLiVRsEZslQT5ktNxAs5wMdsBWpna8mR738SXwlngzOq
Me5axqc83k4m8AwJjIXcxcet38Xipq10tyYofNOjcjPW+af/+Uv2tm23YXvvG7ljBFfaj8RfuyM8ki/6
uR6ga9s29XmyOg+j5Zbxtvkh/AUAAP//yzWLXhACAAA=
`,
	},

	"/": {
		isDir: true,
		local: "/",
	},
}
