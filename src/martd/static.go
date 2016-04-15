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
		size:    2963,
		modtime: 1460723246,
		compressed: `
H4sIAAAJbogA/4xWbW/bthN/LX8K/v0ikv6R5aQZUsBBVwxFsG5YmiL2hgBBUNAybSuRSY+kYmdtvvvu
ThQlxXYwA9bD8R5+d/rdkcMhW1q7NqPhcJEbmy5yuyynaaZWw1uhZKaVMcP35+dn796f9XqbXM7UJl1x
bWfsA5uXMrO5klHMvveCJ64Zf+Db1gKLSl0kbMYtT1jGi2LKs8eECcsXZEI2aCDFJrLL3KS3V398Bjw3
4u9SGMt+/GC/gKcncXs9fRCZjaPwagw670hxMvmanqUnYXwBrrapWgsZYSz2kYVfr8eTkI1Y+OvlJEwY
4TglxXzOogYB2BlhXbzPgs+EjsLf5oMvSorBFbfZMnSA0falt9/gk5JWSDuwz2sB+iFfr4s841iD4Xaw
2WwGc6VXA0AhZKZmYuYxSw0uno3lVmRLLheiUz6PkbTGqMV+Zmfs6MiXs/0coaJZK2nERGxtwrYV6gsH
W86oPiTUwpZasi08w3ov8EHNTy6s04AiLNN5oZSOEEt0yo4rmebAhlUUx+z/7GR7egI/UIhTq8ZW53IR
nZ7HqSmnxr1h2JdeRZRFme8yqI5IcQjGMXO3cBC+9dIy6Cy2bgitzhUR1CT+jtWhl3R8efPX5Q3I+n0v
046JwNGyKLwYv5UUhek6yCgpzC2KvVA8Cf1tWq7WgjLmhRF+DViRQW2+ZYUyolntMfhVGuty2q6T5CuR
MJP/A9cin8MVePoO/gl7FM8JW/PnQvFZ021YWOZ+mDZQEPxFXoa//hCifHQpfehDwTAM3PpHGIkk+EAS
jEoSfIDKdxwdOTSk4J7JCsCRDO7egmiIEyPqlP+4atbdTC7Ikj5gZWC6tcEEQH3aHjDY7P+rXvEK6ido
TwxwCYOs+03v8OG+NnZS14mNDb0GWKgReUroHYOMKFT1XoM3I+AJSuiyLx64dIJq0DiQiy6lagCpd3wH
GmQ8pcQQ8w7vHPhmzbHaybtUT/lUaVtFCw7w1OpSXNTpvDBIQlSu9lHeK6Mkin16rtvJDpIYYa5UtYzL
TBSj19MhCGaiEDAB99fAw2mmDAbcnTJNG/SHBomfz4ibvocRIWTMiFEsl6/40apmtSulwGoT7fuoDcQ4
hdG/sEsY3yd1OgjiGJqxf4ThKRj0ShtL15nfhuo8d2bUwW6iwZK0dhYL+0Pc/mY7gy4IrH52SLFkuLPA
6u/j6y/pmmsjKh8VTTrlQs3X1XLdtC8vrxFksHOpQqSFWkT9P+WjVBtZf+0R6yf0HDfKNpeOWq6xHJB8
B4WroBsqTURMzAkhtzdM7vJ7F6gK8bDLi51P7qO0KhkcaP8WmR/uoxpnHRJ6LMOzCJxctG757VTMC4P+
p/pwcKm10lQ6MGwPVZqVD96kFanXubvbQT4eqBouVS4rB+7kCCenSb4SqrQR9mYCvVDxeU96zbTqTB+f
/Zs7aPAfgnYGVxWv1RWv6fiqki3u1zU6FA5PRnGrGNS81WGonoGUy0XvBafjvwEAAP//2S3x1ZMLAAA=
`,
	},

	"/index.html": {
		local:   "index.html",
		size:    527,
		modtime: 1460723205,
		compressed: `
H4sIAAAJbogA/3ySwUrEMBCGz/UphuChi0tTe1yzvfgKgudsMm4jaVKSqYuI726SLlKF7ekfpl9mPpiK
gUbb31ViQKlTVoIMWexHGUiDxtELvnQSw6+QOHn9WeCogpkIYlBHpqxBR817ZL3gy4cVk8uKc7ivWX7N
do2cJnS6ZmIK2L9KQ8adSy2G0LPdU37wIQMM0mmLcITi1MT5VDP1yPbwNjtFxrtaS5K7r8xXyrvoLTbW
n5d2GbOxmMEDZDAF+7/9O0XOi3HaX5qI9GJG9DPVZejv/uvu27bdDdu/vok7JHClvSX+3B1gS37RL7mH
rm3bXJfO6jqCL6dMpy3/wU8AAAD//3vDFHMPAgAA
`,
	},

	"/": {
		isDir: true,
		local: "/",
	},
}
