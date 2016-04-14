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

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
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

func (dir _escDirectory) Open(name string) (http.File, error) {
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
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
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
		size:    2985,
		modtime: 1460628608,
		compressed: `
H4sIAAAJbogA/4xWbW/bthN/LX8K/v0ikv6R5Tx0LeCgK4YiWDesTRF7Q4AgKGiZtpXIpEdSsbM23313
J4qSYjuYAevheI8//u6o4ZAtrV2b0XC4yI1NF7ldltM0U6vhjVAy08qY4bu3b8/P3p33eptcztQmXXFt
Z+w9m5cys7mSUcy+94JHrhm/59vWAotKXSRsxi1PWMaLYsqzh4QJyxdkQjZoIMUmssvcpDef//gE+VyL
v0thLPvxg/0Cnh7FzdX0XmQ2jsLPY9A5I8XJ5Gt6np6E8QW42qZqLWSEsdgHFn69Gk9CNmLhr5eTMGGU
xykp5nMWNRmAnRHWxfsk+EzoKPxtPviipBh85jZbhi5htH3u7Tf4qKQV0g7s01qAfsjX6yLPOGIw3A42
m81grvRqAFkImamZmPmcpQYXT8ZyK7IllwvRgc/nSFpj1GI/s3N2dOThbD9HqGjWShoxEVubsG2V9YVL
W84IHxJqYUst2RaeYb0X+KDmjQvrNACEZTovlNIR5hKdsuNKpjmwYRXFMfs/O9mensAPFOLUqrHVuVxE
p2/j1JRT494w7HOvIsqizHcZVEekOJTGMXO3cBC+9tIy6Cy2bphaXStmUJP4O6JDL+n48vqvy2uQ9bEn
oCUKBcgulbGjn96cn532vaZ2/ATmlkXhxbiDUhSm6zajUrHiKPZC8Sj0t2m5WgvCgRdG+DXgSgaIfcsK
ZUSz2mPwqzTW5bSNnuQrkTCT/wPXIp/DFdh7Bv+EPYinhK35U6H4rOlBhJu5H4IBxAR/kZfhrz+EKB9c
Se/7ACOGgVv/CCORBB9IglFJgg+wHx1HRy4bUnDPZAXJkQzu3oLIiXMk6mzKcdXCu5VckCVta2Vguthg
AaA+bY8dHAH/q17xCuonaE+8cAWDrLunt/hwVxs7qevPxoZeAwRqRJ4SescgIwpVvdfJmxHwBCV02RcP
XDpBNX5ckosupeoEUu/4FjTIeEqFYc47vHPJN2uO1U7epXrKp0rbKlpwgKdWl+KiLueZQRGicrWP8l4Z
JVHsy3MzgOygiBHWSqhlXGaiGL2cGUEwE4WAubgfA59OM3sw4O7sadqgPzRI/HxG3PQ9jBlCxYwYxXL5
gh8tNKuzKgVWm2jfpjYpxikcCAu7hKF+UpeDSRxDM/aPMDwFg15p59J15g+nus6dGXWwm2iwJK3zxsKp
Ebf3bGfQBYHVTy5ThAzPG1j9fXz1JV1zbUTlo6JJBy7UfImW66Z9dXmNIIPzTBUiLdQi6v8pH6TayHq3
R6yf0HPcKNtcOmq5xnKJ5DtZOATdUGkiYmFOCLW9YnKb37lAVYj7XV7sbLmP0kIyOND+LTLf30V1nnVI
6LEMv1Dge0brlt8OYl4Y9D/WnwyXWitN0IFhe6jSrLz3Jq1Ivc7d3Q7y8QBquFS5rBy470n4nprkK6FK
G2FvJtALFZ/3lNdMq8708dW/eoIG/yFoZ3BV8Vpd8ZKOL5Bscb/G6FA4/F6KW2BQ81afSPUMpFoues84
Hf8NAAD//3qDczWpCwAA
`,
	},

	"/index.html": {
		local:   "index.html",
		size:    528,
		modtime: 1460628608,
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
