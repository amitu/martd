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
	isDir      bool

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
	f, present := data[path.Clean(name)]
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
		gr, err = gzip.NewReader(bytes.NewBufferString(f.compressed))
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (f *file) File() (http.File, error) {
	return &httpFile{
		Reader: bytes.NewReader(f.data),
		file:   f,
	}, nil
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
	return f.isDir
}

func (f *file) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return local
	}
	return static
}

var data = map[string]*file{

	"/client.js": {
		local:      "client.js",
		size:       2651,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\x8cVmo\xdb6\x10\xfe,\xff\n\xce\x1f\"j\x91\xe5\xa4\x19R\xc0AV\fE\xb1nX\x9ab\xf6\xb0\x00AP\xd0\xd2ٖ#\x93\x1eI\xc56\x9a\xfc\xf7\x1d_\xf4f;A\xf3!\x92\xc8{y\xee\xe1sG\x0f\x87d\xa1\xf5Z\x8d\x86\xc3y\xaet2\xcf\xf5\xa2\x9c&\xa9X\r\xef@\xf0T\n\xa5\x86\xef///\u07bd\xbf\xe8\xf569\xcf\xc4&Y1\xa93rMf%Ou.8\x8d\xc8\xf7^\xf0\xc4$aK\xb6mm\x10Z\xca\"&\x19\xd3,&)+\x8a)K\x1fc\x02\x9aͭ\x8b\xf51\x0e\x1c6T/r\x95\xdc\xdd\xfc\xf5\x19\xf1\xfc\r\xff\x95\xa04y~&\xbfa\xa4'\xb8\xbb\x9d.!\xd5\x11\ro\xc6h\xf3\xce\x1aN&_\x93\x8b\xe4,\x8c\xae0\xd46\x11k\xe0\xd4\xe4\"\x1fH\xf8\xf5v<\tɈ\x84\xbf\u007f\x9a\x841\xb18έa>#\xb4A\x80~\n\xb4\xcf\xf7\x19X\x06\x92\x86\u007f\xcc\x06_\x04\x87\xc1\r\xd3\xe9\"\xf4\x80\x8d\xefK\xef\xb8\xc3\xdd\xc0/@6\xf8\x17)D\x9f\xb0[J\x05\xf2\xd0\xf7\xa3\xe0\x1a\xb8\x1e\xe8\xdd\x1a\x8c\x1f[\xaf\x8b<e\x86\xbf\xe1v\xb0\xd9l\x063!W\x03\xac\x00x*2\xc8\xeaz\xb9\xc4\x10;\xa5\x99\x86t\xc1\xf8\x1c:\xd4\xd7\xf5Y\xab\xb1\xb1\"\xbf\x92\vrrR\x1fE\xfb\x9d\x1aC\xb5\x16\\\xc1\x04\xb6:&[Wq\x05\x9bg\x96[\xbb(A\x97\x92\x93-\xbe\xe3~/\xa8\x93\xaa_|Zo\x81\x04.\x92Y!\x84\xa4\x06\v='\xa7nM2TҊF\x11\xf9\x99\x9cm\xcf\xcf\xf0\x0f\r\xa2D\x8b\xb1\x969\x9f\xd3\xf3\xcb(Q\xe5T\xf9/\x93\xf6\xa5\xe7D6/\xf3C\xf5U\x19m\x1e\v\xe3\x94\xf8G8\b\xdf\xfah9t6[\x0f\x03\xad\xaa\xd5 \xa8\x1a\xe0\xbba\xc7~ wN\xb1\xa8\xe5\xb2(\xeaes.\x1c\n\xd55Nm\x01\xa6\x0e\x1aՋ\xf0\x04\xf2۴\\\xad\xc1V\xc7\n\x05\xf5\x1e* E\x1e\xbe\xa5\x85P\xd0\xecV\xdbHT\x9b\x10\x93ԉ\x16\xbbn\xdat\x9a\a\x83\xa6]p\xf7\xe6\xe5\xc1\xb7\xc6O~\xd5˧\xf1\xb1\x9f\x01g+\x18\xd9H\xb1\xfd6YF.\x97\xfd\xae\xe4\xa4FX\xb0Y\xb1\xff\x8e\xe5Ð~\xc1u\x96\a9\xefrS\x01H\xea\xc0\xf7ha\x9d\xa7\x86\x00\x8b\xf9\x80@\x0f\xbe\xd9\xf3\xc7\xe3\u05fbg\x96\xb0\xa9\x90\xdae\v^!\\\xcb\x12\xae\xaar^\b\x16\x01.Ա\xb3\xab\x8d\xcd\n\x8d\xea\xf2\xbcD\xad\x1f\x1612\xb5Z\xd6R\xc6S(F\xfb\x92\x0e\x82\f\n\xc0\xb6=\xceA\r\xa7i\r\x93\xf0\xb05\xcc\x0e\x8e\x0f\xdc\xe8\x0fQ*\x1fP~\xd7}\x94v-F\x83\x10+&V9$\xe7{\xfah\xb1\xe9\xc6p\xf2\b;E\x8f\x1dj\x031Jp^\xcd\xf5\x02g\xceYU\x8e\x01qzMh\xffĤ\xb7\xc9NI\xbf\x8d\xa5\x1b\xac\x9e\xbbU\x9d\a\xcdf.\x1cw\xc9ؖ\x88[\xf3O\xe3\x14\x8bڇtТA\xa0\xe5\xceC3\x1c\x99\xf9\x87\xbb\u007f\x8eo\xbf$k&\x15\xb8\x18N\x17\x1d~\x8c\xe5>=\xbe}\x8e\x15R[\x04)\xceWQ@R\x889\xed\xff\xc3\x1f\xb9\xd8\xf0\xeaxG\xa4\x1f\xdb\xf7\xa81\xd69\xf7Z\xf2\x9d\xe4\x81\xe4\a(<ek\xb6+\x04˚\x8c\xa60\xbf\x88\xb5\xbd\xe1r\x9f?\xf8D.\xc5\xf2P\b\ag\\gi1\x19\xbc\xd2\xef-\xf5.\x1fh\x85\xb3J\x89M\x95\x9a\xdb\x16\xeff)[q;\x8cՋA\xffcu\x85}\x92RHK\x1d:\xc6U\xa9\x8eȘ,k\x97V\xa6^\xe7\xe9\x1f\xaf\n\xf0\x15\xd6̖\v\xe9\x02\xf8\xdfFx\xbfO\xf2\x15\x88RSӌ1\x8a\xdf\t\xf8Hy\xcdxꌛ\xba\xfa7g\u007f\xf0\x03I;\x93jO|{\xbcY\xa5\xbf\x1d\xd4\xdc\xd2Q\xabdۓ\xeeb\xaeF\x9bE|\xd5{1C\xef\xff\x00\x00\x00\xff\xffv\xc9)j[\n\x00\x00",
	},

	"/index.html": {
		local:      "index.html",
		size:       617,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\x84\x92\xcfN\xc4 \x10\xc6\xcfݧ \xc4C7n\xa0ۣ\xb2\xbd\xf8\n&\x9eY\x18[\x1a\xfeT\xa0\xae\x1b\xe3\xbb\v\xb41\x1aSMHf\x98|\xf3\xf1\x9b0l\x88Fw\xbb\x8a\r\xc0e\x8a\x15\x8b*j\xe8\f\xf7Q\"\t\xc61\xbaT\x92\x86\xae\"vv\xf2Z\xc4Ax5E\x14\xbc8aJ\xf9\xc8\xdfH\xef\\\xaf\x81O*\x10\xe1L\xa9Q\xad\u0381\x8e/3\xf8+=\x92c:\xeb\x8d\x18e\xc9\x18p\xc7\xe8b\xf5\xcbUh\x056nirZ\xdd\xd48\x03\xe1=\xe1\xd3\x04V֘M\x1e\xba'\xae\xa2\xb2}\xc9\xd9\xe0;\xbc\xbf\xdfe\xf9+\xf7h\xe0Vj@'T\xe6$a>\xd7X\x1c\xf1\x015\a\xf4<[\x11\x95\xb3\xb5\xe4\x91\xef\xdfsK%\x9c\rN\x03ѮO\xc2\x16\xa5\x86d\x9d\x1d\xabm\x00\x8cnQ\xf6H\x01\xff\xa0H-\x1f+\xccEY\xe9.$@|T\x06\xdc\x1c\xeb\xe2\xf8Ű\xbe\xff\x0f\xc0\xf6L\xed\xf6L\x9b\xdc\x0f\xed\x1d\xfa\x8b}\xa1/\xf1\x80ڦir^*\xdf~\x88\xd1eI\xd2Ҕ\r\xfb\f\x00\x00\xff\xff\x19\x14\x1e\xa6i\x02\x00\x00",
	},

	"/": {
		isDir: true,
	},
}
