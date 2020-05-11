// Code generated by go-bindata. DO NOT EDIT.
// sources:
// templates/00_imports.tmpl
// templates/01_initialize_collections.tmpl
// templates/02_collection.tmpl
// templates/03_collection-structure.tmpl
// templates/04_collection-builder.tmpl
// templates/05_model.tmpl
// templates/06_primary.tmpl
// templates/07_fielder.tmpl
// templates/08_single-relationer.tmpl
// templates/09_multi-relationer.tmpl

package bintemplates


import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}


type asset struct {
	bytes []byte
	info  fileInfoEx
}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

type bindataFileInfo struct {
	name        string
	size        int64
	mode        os.FileMode
	modTime     time.Time
	md5checksum string
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) MD5Checksum() string {
	return fi.md5checksum
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataTemplates00importstmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xaa\xae\x4e\x49\x4d\xcb\xcc\x4b\x55\x50\xca\xcc\x2d\xc8\x2f\x2a\x29\x56\x52\xd0\xad\xad\xe5\xaa\xae\xd6\x55\xc8\x4c\x53\xd0\xf3\x84\x08\x82\xc5\x20\x0a\x14\x34\xb8\xaa\xab\x8b\x12\xf3\xd2\x53\x15\x54\xa0\x22\x56\xb6\x08\x85\xb5\xb5\x0a\x0a\x0a\x0a\x4a\xd5\xd5\x50\xc9\xda\x5a\x25\xae\xea\xea\xd4\xbc\x14\x90\x11\x9a\x60\x73\xa1\x1c\x18\xbb\xb6\x16\x10\x00\x00\xff\xff\x08\x20\xfd\xe4\x84\x00\x00\x00")

func bindataTemplates00importstmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates00importstmpl,
		"templates/00_imports.tmpl",
	)
}



func bindataTemplates00importstmpl() (*asset, error) {
	bytes, err := bindataTemplates00importstmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/00_imports.tmpl",
		size: 132,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588936995, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates01initializecollectionstmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xa4\x90\x31\x93\xda\x30\x10\x85\x7b\xfd\x8a\x17\x2a\xc8\x04\xd3\x93\xa1\x02\x0a\x1a\x48\x41\x97\xc9\x30\x8a\xbd\x26\x9a\xc8\x2b\xcf\x5a\x90\x70\x3b\xfa\xef\x37\x02\x0e\xbb\x38\xae\x39\x15\x1e\x7b\x9f\xdf\xea\x7b\x4f\xb5\xa2\xda\x31\x61\xe4\xd8\x45\x67\xbd\x7b\xa1\x43\x19\xbc\xa7\x32\xba\xc0\xdd\x28\x25\x33\x9b\x61\x19\x2a\xc2\x91\x98\xc4\x46\xaa\xf0\xfb\x02\xa6\x93\x04\x9e\xde\x67\x41\x0a\xac\x76\xd8\xee\xf6\x58\xaf\x36\xfb\x22\x7b\xf6\x7f\x5c\x87\xda\x79\xc2\x3f\xdb\x0d\xcc\x36\xce\xb3\xac\x3a\x45\x74\x0d\x75\xd1\x36\x2d\xa6\x29\x99\xd6\x96\x7f\xed\x91\xa0\x5a\xfc\xb8\xbd\x6e\x6d\x43\x29\x19\xa3\x1a\xa9\x69\xbd\x8d\x99\xb3\x69\x83\xc4\x6e\x84\x22\x2b\xf1\xd2\x12\x06\xc0\x9b\x47\x0a\x41\x7d\xe2\x72\xac\xea\x6a\x14\xeb\xff\x91\x84\xad\x5f\x06\x8e\x92\x7f\x96\x94\x4a\x7c\x2d\x1f\x9f\x45\xaf\xa8\x12\x57\x29\x4d\x40\x22\x41\x8c\x39\x5b\x79\x72\x43\x87\x9f\xbf\xde\x57\x4c\x4e\x78\xe8\x07\xcb\xbe\x51\x0c\xda\x45\x1d\x04\xad\x84\xb3\xab\xa8\xc2\x80\xc6\x64\xf6\x27\x0b\x3e\x13\x69\x4c\x22\xb7\x5c\x13\xa8\x01\x70\x25\x38\x7c\x83\x1b\xf4\x36\x5f\x40\x2c\x1f\x9f\xf5\xda\xdd\x9d\xf9\xb8\x3a\x6f\xc3\x62\xe8\xff\x88\xef\x0d\xe4\xfb\xd5\xf6\x65\x01\x76\x7e\xb0\x2e\x1f\xa1\x78\x12\xce\xfa\x63\x9c\x4c\xff\xbc\xcb\xec\xbc\x49\x46\x15\xc4\x15\x52\x7a\x0d\x00\x00\xff\xff\x79\xca\x66\xcc\xc7\x02\x00\x00")

func bindataTemplates01initializecollectionstmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates01initializecollectionstmpl,
		"templates/01_initialize_collections.tmpl",
	)
}



func bindataTemplates01initializecollectionstmpl() (*asset, error) {
	bytes, err := bindataTemplates01initializecollectionstmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/01_initialize_collections.tmpl",
		size: 711,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588869781, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates02collectiontmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x8e\xbb\x4e\x03\x31\x10\x45\xfb\xfd\x8a\x2b\xf7\x71\x7a\x5a\x42\x41\x41\x42\xb1\x3f\xe0\xac\x6f\x82\x85\x1f\x2b\x7b\x56\x08\x8d\xe6\xdf\x51\x04\x12\x11\x05\xdd\x68\xce\x39\xd2\x55\x8d\xbc\xa4\x4a\xb8\xa5\xe5\xcc\x45\x52\xab\xce\x6c\xda\xef\xf1\xd8\x22\x71\x65\x65\x0f\xc2\x88\xf3\x27\x2a\xb7\xde\xea\xee\xe7\xd7\xba\xc7\xe1\x84\xe3\x69\xc6\xd3\xe1\x79\xf6\xb7\x66\x7e\x4b\x03\x97\x94\x89\x8f\x30\xee\xe2\x20\x0f\x37\xac\x0a\x49\x85\x43\x42\x59\x61\x36\xad\x61\x79\x0f\x57\x42\xd5\xbf\x7e\x9f\xc7\x50\x68\x36\x4d\xaa\xc2\xb2\xe6\x20\x84\x4b\x65\x6d\x5d\x86\x83\xff\x4b\x7e\x37\xef\x86\xf4\x6d\x91\xad\xd3\xc1\xbf\xb4\xc8\xfc\x8f\x7b\xde\x52\x8e\xec\x77\xa6\x2a\x58\x23\xcc\xbe\x02\x00\x00\xff\xff\xaf\xc9\x98\x0d\x10\x01\x00\x00")

func bindataTemplates02collectiontmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates02collectiontmpl,
		"templates/02_collection.tmpl",
	)
}



func bindataTemplates02collectiontmpl() (*asset, error) {
	bytes, err := bindataTemplates02collectiontmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/02_collection.tmpl",
		size: 272,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588867762, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates03collectionstructuretmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x54\x41\x6f\x13\x3d\x10\xbd\xef\xaf\x98\xaf\x87\x76\x37\xda\xcf\xe5\x1c\x29\x1c\x1a\x10\xe2\x50\x24\x40\xe2\x52\x55\xc8\xf1\x4e\xda\x11\x8e\x77\x71\xbc\x21\xc5\xf2\x7f\x47\xf6\x38\xc9\x26\xda\x94\x4a\x70\x24\x07\x47\x63\xcf\xbc\x37\xf3\xe6\x25\xde\x37\xb8\x24\x83\x70\xa1\x5a\xad\x51\x39\x6a\xcd\xff\x6b\x67\x7b\xe5\x7a\x8b\x17\x21\x14\xc5\xb2\x37\x0a\xc8\x90\x2b\x2b\xf0\x05\x00\xc0\xd7\x43\xee\x7b\x43\x8e\xa4\xa6\x9f\x68\xd7\x30\x03\xd9\x75\x68\x9a\xf2\x4c\x42\x9d\x60\x38\xf2\x5e\xcc\xf7\x49\xe2\x8b\xb4\x24\x17\x1a\x3f\xc8\x15\x86\x50\x15\xa1\x28\xae\xaf\xe1\xb9\x1c\xa0\x35\xb8\x47\x84\x03\x13\xf4\x6b\x6c\xc0\xb5\xf0\xbd\x47\xfb\x14\x8b\x6f\xdb\x06\xb5\xc8\xf9\xab\x14\x14\x1b\x69\x9f\xc7\x9d\x1c\xbf\xf2\x6d\x51\xb8\xa7\x0e\x61\xec\x09\x58\xad\x2c\x4d\xa2\x84\xc9\x4a\x76\x1d\x99\x07\xee\xe0\x73\x4a\x28\x86\x52\xfe\x5e\x83\xd2\x7b\x5a\x82\x78\xbb\x75\x68\x8d\xd4\xf3\xd6\x38\x1b\x73\x6d\x08\x0a\x26\x6a\x1f\x8a\xc3\x8b\xf7\x68\x9a\x10\x2a\x40\x6b\x5b\x9b\x1b\x5a\x31\x7b\x1d\x2f\x61\x3a\x03\x25\xde\xa1\x1b\xb4\x55\x5e\x9e\x08\x55\xa5\x32\x5a\xa6\x82\xff\x66\x60\x48\x67\xa8\xf8\xb1\xe8\x7a\x6b\xe2\x5b\xba\x0a\xe9\xf4\x1e\xce\x0d\x02\x21\xc0\x0c\x2e\xc7\x84\xf3\x89\x75\xba\xeb\x90\xa1\x32\xbe\x21\x9d\x3d\xf0\x31\x2d\x53\x59\x94\x0e\x79\xe3\xbc\xde\x65\x6b\x53\x74\xd2\xbd\x60\x8d\xcb\x63\xc2\x4f\xa8\x90\x36\x51\xbb\xf1\xf5\x56\x4c\x53\x36\x0b\x46\x17\x6f\x6e\x6a\x36\xcc\x1a\x84\x10\x93\x53\x89\x4e\x51\x52\xf5\x4d\x4f\xba\x49\x1c\x2c\x57\x74\x5a\x42\xbb\x65\xa0\xbb\x7b\xc6\x4e\xe1\x4e\x64\x8d\xa6\x64\xa2\x0a\x5e\xc3\xab\x81\xd2\xc3\xd2\x19\xac\xe4\x37\x2c\x8f\x10\xea\x61\x6d\xb5\x2f\x8b\xba\x50\x6e\x3e\xee\xdb\x4a\xf3\x80\xbb\x59\x0e\xe8\x27\x0c\x77\x74\x1f\x49\xf6\x9d\x1d\x76\xcb\xe7\x82\x67\x8b\x80\xcd\x82\xc7\x3d\x2b\xf1\xae\xbd\xe1\x00\x42\x88\x6a\xb8\xdf\x51\x43\x0c\x55\xf4\x99\x71\xba\xa3\x0e\x43\x3f\xcc\xdd\xf6\xe5\x96\x80\x1f\xe4\x1e\xa1\xb3\xed\x86\x1a\x6c\xe0\x4a\xb9\xed\x15\xc4\x1f\x10\x6e\xdd\x1f\xd9\x65\xee\xb6\xa5\x8a\x9d\x64\xac\x39\x7f\xd7\xf0\xcf\x46\x2f\xb7\x51\x16\xb1\x86\xbf\x68\xa7\x63\x1d\x47\xad\x94\xff\x2b\x7f\x05\x00\x00\xff\xff\x0d\x51\x43\xd8\xff\x06\x00\x00")

func bindataTemplates03collectionstructuretmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates03collectionstructuretmpl,
		"templates/03_collection-structure.tmpl",
	)
}



func bindataTemplates03collectionstructuretmpl() (*asset, error) {
	bytes, err := bindataTemplates03collectionstructuretmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/03_collection-structure.tmpl",
		size: 1791,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588866333, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates04collectionbuildertmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x58\x4f\x6f\x1b\xb9\x0e\xbf\xe7\x53\xf0\x05\x0f\x98\x71\x10\x8f\xdf\xe1\xe1\x1d\x02\xe4\xd0\xe6\x35\x45\x81\x6e\x36\x9b\x2c\xf6\x52\xf4\x20\x8f\x38\x8e\xb6\x1a\xc9\x2b\x69\x92\x78\x07\xfe\xee\x0b\x52\x9a\x7f\xa9\x93\xa0\x9d\x1e\x7a\x49\x6c\x89\x22\x7f\xfc\x91\x22\x29\xb7\xad\xc4\x4a\x19\x84\xe3\xd2\x6a\x8d\x65\x50\xd6\x2c\xd7\x8d\xd2\x12\xdd\xf1\x7e\x7f\x74\xb4\x5a\x41\xdb\x16\x17\xfd\x66\xf1\x5b\x83\x6e\xf7\x36\x4a\xec\xf7\xa0\x3c\x84\x3b\x84\xbf\x68\x15\xd2\x41\x68\x3c\x4a\x08\x16\x4a\x87\x22\x20\x08\x23\x01\x1f\xb1\x6c\x02\x92\x3e\x92\x55\xe8\xa1\xb2\x8e\xcf\xb6\x6d\x71\x25\x6a\xdc\xef\x6b\x2b\x51\x17\x47\x61\xb7\xc5\x97\x8d\xfa\xe0\x9a\x32\x40\x7b\x04\x00\xbd\x51\x86\x50\x24\xa1\xa3\x08\xfd\xb6\xb4\x5b\x04\x87\xa1\x71\xc6\xc3\x46\xdd\xa3\x49\x50\x3d\xed\x14\x47\x55\x63\x4a\xc8\xa7\xc6\x6e\xb0\x44\x75\xcf\x86\x4e\x5e\x82\xb1\x88\xea\xf3\x05\x9c\x44\xdb\xd1\x5a\x44\x15\x6d\xc2\xba\x48\xf0\x8a\x24\x9b\x80\xbd\x73\xae\x87\x85\xce\x59\x47\x34\x8a\x00\xb6\x2c\x1b\xe7\x50\x82\x6c\x9c\x32\x9b\x31\xaf\xf4\x75\xeb\x6c\x89\xde\xcf\xc5\xfd\xce\xb9\x7c\x11\xed\x3e\x07\x97\x45\x12\xd8\x8b\xf0\xd8\x83\xa5\x80\x95\xd6\x04\x7c\x0c\x60\xab\x09\xa5\xdd\xd1\x99\xe0\x2e\xc2\x63\xbe\xe8\x6c\x14\x17\xc9\xd6\x33\x30\x59\xb8\x83\x69\x1b\x13\x26\x40\x4d\x53\xaf\xd1\x11\x4e\x4e\x2d\x50\xc6\x07\x61\xca\x94\x7b\x5b\x67\xef\x95\x44\x99\x32\x67\x2e\x6c\xb2\x9e\x2f\x20\x57\x26\xfc\xef\xbf\xa7\x91\xdd\xc5\xb3\xb8\xa3\x74\x87\x3c\xde\x13\x65\x3c\xba\xe0\xc1\xe0\xc3\x13\xc0\xa0\x4c\xb0\xec\x92\x0f\xd6\xcd\xce\xdb\x68\xef\xb5\x14\xe8\xa4\x12\xc8\x6b\x11\xca\x3b\xd8\xd2\x5f\x9c\xde\xa5\x88\xf5\x41\x85\x3b\xf0\x48\x36\x51\x42\xa5\x50\x4b\x0f\xca\x30\xea\xf8\x0d\xc3\x5c\xe0\x8c\xe1\x35\xdc\x49\x28\xc1\xbe\x54\x46\xf6\x49\x21\xb4\xa6\xca\xf2\x0b\x17\x9a\x58\x74\x22\xfa\x74\xfd\xea\xe4\x5d\xb0\x63\x07\xe7\xa2\x26\x08\x94\x19\x9f\x3e\x9f\x3c\x31\xfe\x24\x4d\xd8\x1a\xef\x7b\xde\x81\xb3\xf3\x91\x5f\x51\x0d\x0b\xaa\x8a\xb7\xff\x75\x0e\x46\xe9\x74\x78\xc4\x85\x51\x9a\xcf\xf3\xfa\x9e\xff\x26\x2f\xcf\xce\xa1\x16\x5f\x30\x42\x01\x46\x01\x04\x43\xa3\xc9\x47\xd6\x17\xd1\x0c\xdd\x12\x75\x9a\xe2\x7b\x76\x0e\x4e\x98\x0d\x8e\x51\x8e\x4c\x47\x03\x9f\xd4\x67\x38\x1f\x4b\x7c\x52\x9f\x8b\x7c\x6c\x6b\x31\x02\x95\xe0\xd6\xc9\x63\xa3\x74\x0a\xda\x7b\x1c\x2e\xb2\x57\x66\xa3\xf1\x70\xd8\xa6\x51\x9b\x84\x6c\xb5\x82\x0f\x15\x27\x5f\xba\x4a\x1e\x8c\x0d\x50\xd9\xc6\xc8\x98\x93\x8d\xe1\x70\x4d\xeb\x30\x15\x8b\x52\x0b\xef\x63\x14\xff\x10\xba\xc1\x2b\x7b\x83\xbe\xd1\xb3\xb3\xf7\x3d\x72\x81\x78\x25\x09\x18\xee\x81\xf0\xf3\xe9\xef\x8c\xfe\x98\xe8\x14\x8e\x01\x01\x27\xc0\xc0\xfd\xff\x51\x63\x40\x90\xfc\xef\xf0\x95\xb9\x27\x56\x3c\x54\xce\xd6\x43\x51\x7a\x21\x16\xf3\x68\x8b\x80\x5e\xbb\xf5\x9d\x54\xf2\xe2\x8d\x94\x97\x4a\x07\x74\x97\x54\x7a\x40\x48\x19\x5b\x42\x56\xf1\x6a\x06\xa9\x9e\xfe\x40\x9c\x53\x93\x79\x34\x04\x27\xa3\xb5\x14\xfc\x17\x74\xa4\x91\xa6\xf7\xea\xa0\xce\xc5\x84\x81\xbe\xce\xb1\xb9\x38\x72\xf9\x54\x99\x7b\x6f\xb9\x38\xf7\x0d\x2f\x8b\x01\xcc\xe6\x17\x36\xd2\xde\x79\xea\x03\x8d\x2c\xa7\x5d\x76\x14\x45\xa1\x4c\x40\x57\x89\x12\xdb\xfd\xb7\xbb\x3e\x51\xde\x69\x2d\x8a\xe2\x6b\xef\x8f\xda\x36\x95\xa6\x7f\x3b\xd4\x82\x2f\xf5\xd9\x39\x14\x37\xe9\x8b\x87\xfd\x9e\xeb\x81\x29\x75\x23\xb1\x6d\x7b\x31\x1a\x60\xe3\x62\x97\xb1\x59\xb7\x95\x75\xbd\xab\x4f\xfe\x8b\xfd\x7e\x54\x5c\xde\x68\x6d\x1f\x90\x5a\x4b\x67\x24\xd8\x4e\xd9\x19\x09\xc0\x12\xda\x76\x40\xd4\x5d\x35\x3e\x6b\xc0\x6e\x69\x51\xe8\xc1\xe0\x65\xea\x91\x19\xc4\x91\xdc\x8f\x3a\x27\xe9\x4e\x1e\x27\x54\xc9\xd2\x60\x7f\x6e\x2c\x7b\x72\x96\x23\x16\x97\xfb\x7d\x4e\x2b\xc1\x7e\xb4\x0f\xe8\x2e\x44\x8d\x7a\xba\xdd\xa1\xa6\x70\xc7\x04\xf8\xf6\x48\x27\xd3\xf9\xf1\x24\x32\xc7\xa7\xf0\x94\x9a\x43\xd1\x6f\x5b\x34\x32\xbd\x54\x3e\xaa\x5a\x05\xf0\x18\x22\x77\xb5\x78\x54\x75\x53\x8f\xc6\x40\xbb\xfe\x13\xcb\xe0\xd3\x79\x94\xb0\xde\xb1\x24\x0f\x09\x69\xb8\x3e\x25\x4d\x37\x93\x9e\xa0\xaa\x51\xad\xe0\x97\x03\xdc\x09\x2a\x8d\x0e\x85\xdc\x81\x54\x55\x85\x0e\x4d\x00\x7e\xba\xd8\x0a\xb6\x62\xa3\xcc\x0f\x09\x0b\xbb\x94\x6b\x76\x8c\x27\xcb\x6f\xa7\x77\xa4\xe2\x60\xe5\xf8\xb5\xaa\x28\x80\x3d\x6d\xb1\x74\x38\x6e\x79\x99\x07\xcb\xdb\x05\x7c\x08\xe0\xc5\x8e\x53\xd1\x7f\x51\x5b\x10\x1e\x6a\x61\x76\x89\xd3\x6c\xd4\x0e\x1c\x6e\xad\x57\xc1\xba\x1d\xa9\x5f\x63\x45\xdd\x61\x8d\x1b\x65\x0c\x3d\x62\x86\x5c\x8e\xc2\xdc\x5a\x21\x8b\x30\x32\xf8\x4f\xf7\xa6\xf4\x74\x61\x84\x07\x5b\xab\x10\xf8\x20\xd5\xef\x4e\xac\xd4\xa2\xf1\x58\xfc\x54\xc1\x8a\xd8\xf2\xc8\xd8\xf7\x86\x6b\xa2\xe4\x60\xc0\xae\xc5\x06\x6f\xd5\xdf\xd8\x17\x8a\xc1\x03\xfa\x88\xe0\x69\x73\xd9\xe7\xbf\xa8\xf9\x65\x64\xab\x21\xef\xd3\x45\x38\x4c\xdf\x98\xba\x8e\x36\xa6\x70\x6c\xc8\x56\x31\x37\x23\xdc\xeb\x61\x83\x68\x9d\x3f\xe7\x47\x0f\xf3\x6d\xe7\xea\x77\x92\xf9\x95\xa2\x67\x09\xbd\x8a\x45\x62\x5c\x7b\x9f\xd2\x1a\xeb\xc8\xcf\xcc\x59\x74\x82\x9d\x4d\xfe\xcc\xe0\xed\x2b\x65\xcf\xcc\x1d\xdc\xa0\xfa\x01\x6b\x68\x58\x7c\x87\x89\x10\xaa\x0d\xfd\x03\x70\xb5\x82\xdf\x07\xb1\x5a\xec\x60\x8d\x20\xa0\x16\xdb\xad\x32\x9b\xe2\x96\x7f\xda\x49\x43\x9b\x87\x07\xd4\x9a\xfe\xb3\x78\xe6\xe1\x0a\x1b\x67\x0d\xb7\xd2\xbc\x6b\x37\xd6\x91\x52\xb2\x36\x3e\x3c\x91\x99\x3f\xe8\x10\xda\x3c\x81\x9e\xd1\xea\x26\x8a\x0e\x0e\x33\xab\x15\xdc\x5a\x17\x06\x42\x3d\x7d\xeb\x9f\xd3\xc1\x4e\x7e\xc0\x82\xb7\x3b\x4a\x59\xd1\xe8\x24\x43\xc5\xd3\x3a\x89\x0e\x89\xbf\x12\x8d\x24\x56\xe1\x83\x89\xab\x5c\xbd\x49\xa1\xc4\x6e\x93\x2c\x0a\x29\x21\x5b\x66\x5d\xad\xee\xe3\x08\x86\x58\x54\x05\x16\x90\x2d\x95\xcc\x0a\x8e\x5d\xd4\x64\xab\x71\xbc\xb9\x55\xe3\x10\x77\xb5\x31\xaa\x52\x25\xff\x78\x91\x24\xc9\x2e\x95\x71\x3e\x3e\x99\xa1\x86\x9c\x21\x99\x6e\x7a\x2a\xae\x9d\xaa\x85\xdb\x31\x67\x3c\x3e\x8d\x06\xbd\x08\x8f\xa6\xbc\x94\x80\x7b\x1a\x05\xe8\x26\xc6\xad\x82\x48\x14\x6b\xdd\x8d\x5c\xb0\xec\x36\x3a\x45\x68\x64\x3a\x94\x3e\xcd\xfc\x2d\xd0\xba\xf0\x03\x12\x64\xa4\xe6\xf0\xb4\x93\xd0\xfe\x13\x00\x00\xff\xff\x67\x52\xc9\x0b\xb9\x15\x00\x00")

func bindataTemplates04collectionbuildertmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates04collectionbuildertmpl,
		"templates/04_collection-builder.tmpl",
	)
}



func bindataTemplates04collectionbuildertmpl() (*asset, error) {
	bytes, err := bindataTemplates04collectionbuildertmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/04_collection-builder.tmpl",
		size: 5561,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588870486, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates05modeltmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x91\xcd\x4a\xc3\x40\x14\x85\xf7\x79\x8a\x43\xf6\x49\xf6\x6e\xad\x82\x0b\x5b\xd1\xbc\xc0\xd5\x39\x89\x17\xe7\x27\x4c\xa6\x48\x19\xf2\xee\x32\x4d\x09\x05\xb5\xbb\xcb\xf9\xbe\x73\x36\x37\xe7\x06\x86\x83\x7a\xa2\x76\xc1\xd0\xd6\x68\x96\xa5\xea\x3a\xdc\x07\x43\x8c\xf4\x8c\x92\x68\xf0\x7e\x82\xe7\x31\x06\xdf\x5d\xb2\x10\x5b\xec\x0e\xd8\x1f\x7a\x3c\xec\x9e\xfa\xb6\x74\xfa\x4f\x9d\x31\xa8\x25\xbe\x65\xbe\x2a\x4b\xba\x2b\x38\x67\x24\x75\x9c\x93\xb8\x09\xcb\x52\x55\x93\x7c\x7c\xc9\x48\xe4\xdc\xbe\xac\xe7\x5e\x1c\x0b\x29\x2a\xdd\x64\x25\x11\xb5\xba\x29\xc4\x34\xd7\x68\xf1\x8b\x4d\x51\x9d\xc4\xd3\x85\xe5\xac\x03\xda\x47\xa5\x35\x8c\x6b\x70\xe5\x0e\x6b\xbe\xb9\x0d\xe8\xcd\xd6\x7a\x53\x3f\x5a\xbe\xd2\x4a\xd2\xe0\xff\xa8\xcf\x67\xa1\x89\x9b\xf1\xcf\xd0\xf3\xd1\x26\xbd\xb1\xe3\x0a\xbf\x3d\x73\xbe\xca\x27\x7e\x02\x00\x00\xff\xff\xeb\xde\xbe\x97\xa0\x01\x00\x00")

func bindataTemplates05modeltmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates05modeltmpl,
		"templates/05_model.tmpl",
	)
}



func bindataTemplates05modeltmpl() (*asset, error) {
	bytes, err := bindataTemplates05modeltmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/05_model.tmpl",
		size: 416,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588981718, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates06primarytmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x96\xcd\x6e\xdb\x38\x10\xc7\xef\x7e\x8a\x81\x91\x8d\xa5\x85\xcd\xec\x5e\x53\xf8\xd0\x06\x68\x10\x14\x0d\x82\x24\x68\x81\x5e\x0a\x56\x1a\xd9\x44\x24\x52\xa1\x68\xa5\x06\xc1\x77\x2f\x48\xca\x12\x65\x5b\xb6\x1b\xa4\xb7\xe6\x12\x83\x9a\xf9\xcf\x6f\x3e\xf8\xa1\x75\x8a\x19\xe3\x08\xe3\x52\xb2\x82\xca\xf5\x18\x66\xc6\x8c\xb4\x9e\xc1\x99\x58\x29\xb8\x9c\x03\x71\x2b\x17\x17\x70\x25\x8a\x92\xe5\x08\x8a\x15\x08\xc9\x12\x93\x27\x60\x19\x68\x4d\x6e\x69\x81\xc6\x00\x2b\xca\x1c\x0b\xe4\xaa\x82\x82\x96\x25\xe3\x0b\xf2\x59\xa4\x98\x03\xe3\x0a\x65\x46\x13\x24\xa3\x9a\x4a\xf8\xbe\xf5\x79\x0e\xe7\xad\x88\x36\x23\x1b\xea\x16\x57\x52\xf0\x2b\x91\xe7\x98\x28\x26\xb8\xfd\x78\x82\x3e\x14\xa8\x96\x22\x25\x56\xe2\x1e\xd5\x4a\xf2\x0a\xd4\x12\x81\x5b\x77\x91\xb9\xdf\x49\x2b\x0a\x99\x90\x6e\x69\xd2\x86\x9f\x90\x51\xb6\xe2\x09\x44\x5a\x93\x7b\x4c\x90\xd5\x28\x8d\x81\x7f\x5b\x83\x78\x2f\x5a\x14\x43\xa5\x24\xe3\x0b\xd0\x23\x00\x00\xe9\x62\xc3\x58\x6b\xd2\x37\x34\x66\x3c\xf2\x19\xde\x54\x77\xbe\xe0\x9f\x70\xfd\x0d\xa5\x08\xb3\x7b\x5e\xa1\x5c\x0f\xe6\x76\x0c\x70\x5b\x39\x8a\xe1\x87\x10\x79\x1f\x4d\x6b\xd2\x58\x91\x9b\xca\x5a\x19\xd3\x80\x5d\xa3\xea\xfc\xbf\xd0\x7c\x85\x6f\x87\xb6\xab\x1d\xc5\x9d\x8c\x36\x3b\x8c\x9d\x0e\x09\x88\xbd\xdc\x3e\x5e\x9b\xc9\x9f\x64\x6e\xf5\x8f\x70\x6f\x48\xaf\x51\xf5\x8a\xfb\x10\x8a\xbd\x1d\xe3\xc3\x6e\x5d\x6b\x5f\x85\x8e\x31\x06\x94\x52\xc8\x06\x55\x6b\x96\x41\x30\x02\x77\xc2\x99\xba\xad\x6e\xbf\xb3\x0c\xbc\xc2\x7c\x0e\x9c\x6d\x86\xc7\x7b\x1e\xec\x0a\x38\xfb\xd6\xba\x29\xc9\x66\xc9\x34\xc1\x91\xa7\xbd\x50\x53\x10\x4f\xf6\xac\x71\x31\x49\x14\xa8\x3e\xae\x4b\x9b\xe2\x3b\x6b\xf0\x3b\x10\xf5\x10\xc2\xb1\xcc\x0d\x60\x5e\xe1\x11\xaa\x0f\xb4\xc2\xd7\x92\x9d\x1f\x40\x6b\xea\xa2\xb5\xa4\x7c\x81\x70\xf6\x22\x69\x59\x62\xea\x8e\xe1\x8d\xd0\x57\xbf\x66\xc3\x57\x27\x51\x6f\x54\x5e\x03\x3b\x50\x2d\x63\xce\x1d\xac\x31\xf5\x9e\x8e\x9a\x3d\x85\x7e\x9f\x2b\x94\x9c\x2a\xec\x73\xdb\x6b\xc5\xdd\x24\x74\xf3\x1d\x94\x33\xb0\x67\xf3\x82\xd5\xc8\x21\x63\x98\xa7\xa4\x93\xc4\x67\x88\x72\xe4\x43\xd2\x31\xfc\x7f\x78\xb4\x18\x4f\xf1\xe7\x20\xd8\x7f\xfb\xaa\xc4\x32\x77\x1b\x1e\x98\x1a\xfb\xa7\xb0\x28\x6d\x24\xad\x7b\xc6\xdd\xac\xf8\x4d\x19\x07\xc2\xce\xb0\xd7\x82\x9e\x6b\x37\x33\x56\x3b\xf0\x9b\xf9\x76\x87\xd1\x4f\x17\xdb\x5a\x1f\x62\x9b\x81\x6b\xf0\xb1\xad\xbc\x85\x52\xbd\x30\x95\x2c\x7d\xb5\xad\x72\x50\x7a\xdb\xd8\xb8\x3d\x7f\x9a\x09\xef\xfa\x1e\xce\xf8\xc0\xb4\x24\xb4\x42\x8b\xdf\x3a\x19\x73\xf9\xb7\x4d\x07\xdb\xb4\xb5\x33\x9b\x5d\xdb\xfc\xdb\xb3\x4f\x1f\x12\xca\x79\x50\xb1\xd3\xee\x62\xe7\x16\xb2\xed\xe4\xde\xe8\xb8\x3b\xa8\x22\xb7\xf8\x92\x45\x49\x4e\xab\xca\xdf\x79\x1f\xed\x1e\xb7\x19\x4e\xed\x43\x54\xd4\x2c\xc5\x14\x18\xaf\x69\xce\x52\x3f\x3e\x97\x30\xf9\xe7\x71\xd2\x3e\xd9\x9a\xd7\xaa\x3f\x1c\xdc\x6a\x61\x75\xbc\xd5\x78\xda\x56\xc3\xf9\x4e\xfb\xf8\x01\xa1\x2b\x82\x7f\xee\x36\x65\xfa\x15\x00\x00\xff\xff\xa2\xd5\xa5\xc6\x11\x0b\x00\x00")

func bindataTemplates06primarytmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates06primarytmpl,
		"templates/06_primary.tmpl",
	)
}



func bindataTemplates06primarytmpl() (*asset, error) {
	bytes, err := bindataTemplates06primarytmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/06_primary.tmpl",
		size: 2833,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588984338, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates07fieldertmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xe4\x57\x5f\x8f\xe3\x34\x10\x7f\xcf\xa7\x18\x55\x7b\xbb\x49\xd5\x4d\x8f\xd7\xa2\x7d\xe0\x40\xc0\x3e\x70\x87\xe8\x0a\x24\x10\x42\xbe\x64\xb2\xb5\xce\x71\x82\xed\x66\x39\x59\xfe\xee\xc8\x4e\x9a\xe6\x6f\x93\xb2\x54\xec\x8a\x3c\xc6\xf3\xe7\x37\xf3\x9b\x19\x8f\xb5\x8e\x31\xa1\x1c\x61\x91\x50\x64\x31\x8a\x05\xdc\x1a\xe3\x69\x7d\x0b\x57\xd9\x5e\xc1\xe6\x0e\x42\xf7\x67\xbd\x86\xaf\xb3\x34\xa7\x0c\x41\xd1\x14\x21\xda\x61\xf4\x09\x68\x02\x5a\x87\xef\x49\x8a\xc6\x00\x4d\x73\x86\x29\x72\x25\x21\x25\x79\x4e\xf9\x63\xf8\x6d\x69\x14\x28\x57\x28\x12\x12\x61\xe8\x15\x44\xc0\x1f\x3d\x81\x3b\xb8\xae\x0d\x69\xe3\x59\x77\xdf\xa1\x72\xa7\xbf\xa2\xc8\x7e\x26\x6c\x8f\xf3\x1c\x48\x2f\xd9\xf3\x08\x7c\xad\xc3\x9f\x30\x42\x5a\xa0\x30\x06\x96\xb5\xf5\xa0\x6f\xd8\x77\xb1\xc3\xf2\x60\x74\xab\xc4\x3e\x2a\x65\x02\xf0\x6b\xd3\xda\xac\x00\x85\xc8\x44\x00\xda\x03\x00\x90\x4f\x54\x45\x3b\x70\xda\xe1\x3d\x8f\xf1\xaf\xdf\xde\xfe\x5e\x9d\x69\x2d\x08\x7f\x44\xb8\x2a\x6d\xdb\x3c\x3a\x83\xd2\x65\xd3\x4a\x44\x44\x22\x68\x7d\xd5\x50\x37\x66\x03\xeb\xf5\xf1\x67\x89\xd8\x49\xdb\x4f\xa0\xda\x0b\x7e\x3c\xb6\x01\x18\xb3\x02\x4e\x59\xe5\x13\x79\x5c\xdb\x8f\x31\x21\x7b\xa6\x36\x5d\x75\x4e\x59\x15\x87\x0c\xdf\xe3\x53\xe2\x47\x8c\x48\x19\xfe\x90\xc5\xc8\x1c\x44\xeb\x75\x05\x8b\x5c\x64\x05\x8d\x31\x06\xca\x0b\xc2\x68\x5c\xc6\x09\x9c\xa4\xb8\x81\x9b\x37\xf2\x66\xb1\x82\x23\x4c\x3f\x08\x9c\x23\xe3\x95\xec\xdd\xcb\x3a\xc7\x33\x0b\x63\x8a\xb6\x86\xc5\xd3\x84\x7d\xcc\x32\xf6\xc2\x98\xba\x97\xa3\x5c\x95\x4a\x95\x42\x42\x98\xc4\x0b\xb1\x53\xf2\xb2\xbd\x54\x57\xf5\x0c\x9f\x24\xc9\x45\xf8\x1f\x91\xa3\xb5\x1d\x6d\x8d\x38\xc2\x8e\x24\xdc\x75\x7b\xec\x8c\xf6\xba\x60\x67\xb5\x9b\xd8\xab\xc7\xe4\xf7\x44\xee\xc8\x47\x86\xce\xc7\x19\xa4\xce\x99\x94\x7d\xdb\xaf\x65\x5a\x6a\x4d\x13\xa8\xfb\xef\xc7\xcc\x21\xab\xed\xd9\xcf\xdd\x5e\x93\xb5\x70\x67\x93\x5d\xe1\x1c\x1a\xa5\x87\x8e\x3e\x72\xd4\x2f\x96\x3e\x9e\x77\x9f\x15\x6e\x19\x8d\xb0\x25\xd3\xb0\x2d\x95\xa0\xfc\xd1\x1f\x8c\xc2\x98\xa5\xb3\x6f\xcc\x24\xfe\xa0\x0d\xd0\x5e\xec\xc8\xe4\xa8\xd7\xe7\xba\x1b\xf0\x66\xf5\x4e\x8c\xbc\x53\xd7\xd1\x7d\xd9\x1f\xae\x14\x46\xfb\xa6\x6c\x19\x48\x32\x01\x8f\xb4\x40\x0e\xa9\x55\x1d\x6c\xa4\x15\xb4\x6a\x3d\xf0\xda\x9b\xc6\xbf\xdc\x3a\xaf\xaf\x65\x8e\xd7\xd6\x5c\x96\x5f\x2c\xa7\xdb\x0b\x70\xba\x9d\xc9\xe9\x0a\x8a\xd2\xeb\x91\xd8\x00\x7c\x14\xe2\x15\x0c\xc4\x12\xf9\xe0\xd0\x9b\x73\x6f\x36\xbb\xbf\x73\x5d\xcd\x99\x91\x16\xc0\x0a\xb2\x4f\x36\x60\x87\x24\xf4\x6b\x1f\x0f\x9f\x73\xcb\xc2\x97\xf6\xf8\x7c\x60\xc5\x24\xac\xe1\xf4\x34\xc0\xd9\x27\xd0\xe1\xd5\x43\x15\x50\x09\x3c\xe3\xb7\x79\x25\x58\xc2\x9d\x15\xc9\x3b\x22\xf1\x79\xd1\x5c\xcf\x09\xa7\x4a\x6e\x5d\x4a\x4f\x82\xe4\x39\xba\x62\xaa\xec\xfd\x52\xfe\xb1\x58\xe4\x24\x0b\x95\xfa\x3f\x07\x3d\x72\xb7\x5c\x57\x77\x4b\xd1\xc1\x6e\x5a\x8c\x7c\xc5\x14\x0a\x4e\x14\x76\xc1\x3a\x21\xfc\x13\x7c\x86\x7c\x58\x38\x80\x2f\x5a\x45\x56\xf3\x48\x0e\x62\xa0\xac\x9c\x1b\x37\x6a\x87\xdd\x06\x9a\x20\x95\xda\xd6\x1b\xf6\x0c\x6f\xc7\xb2\x35\xd5\x89\xf6\x53\x98\xe6\xd6\xd3\x40\xdd\xf8\x45\x70\x7e\xcd\x58\x7b\x1d\xad\x91\x65\xe0\xac\x0d\x79\x0c\x51\xf3\xea\x3f\x55\xa8\xa7\x77\x93\x6a\x48\x16\x8d\x94\x5b\xb2\x82\x46\x46\xeb\x02\x27\xcd\xe4\x37\xca\xbc\x43\x4a\xd3\xfc\x61\x8a\xb6\x54\x8d\xd9\xfc\x6f\xe9\x9a\x9c\x82\x2e\x63\xcb\xa9\x94\x9d\x4a\xc6\xf2\xd9\xd9\xe8\x17\x57\x35\x31\x5a\x12\xeb\x25\x3c\x7c\xf8\xe6\xc3\x06\x24\x2a\xd7\xd7\x32\x22\x9c\xa3\x00\x22\x0f\x65\x65\x4b\xc9\x45\xb4\x5c\xb7\xb2\xda\x7b\xdc\xc1\xbc\x07\x9e\x8d\x70\xfc\x81\x67\xbd\xd9\x1d\xe6\xe1\xa6\x1e\x35\xd5\xb6\xf3\x46\x2e\xaa\xa5\x61\xe8\xe5\x07\xad\x99\x38\xb8\x52\x5f\xe6\x49\xda\x58\xc3\x2c\xd8\x7a\x09\x7b\xe8\x2f\x61\xe9\xe8\x23\xb5\x46\xfc\x77\x00\x00\x00\xff\xff\x58\x3f\xc9\x9c\x6c\x14\x00\x00")

func bindataTemplates07fieldertmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates07fieldertmpl,
		"templates/07_fielder.tmpl",
	)
}



func bindataTemplates07fieldertmpl() (*asset, error) {
	bytes, err := bindataTemplates07fieldertmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/07_fielder.tmpl",
		size: 5228,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588981567, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates08singlerelationertmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x55\x51\x6b\xdb\x30\x10\x7e\xcf\xaf\x38\x42\xd6\xd8\xc1\x71\xf6\x9c\x91\x97\x15\x36\x02\x5b\x19\x49\xdf\xc6\x18\xc2\x3e\xa7\xa2\xb2\x64\x24\x39\x59\x11\xfa\xef\x43\x92\xed\xc4\x76\xd7\x76\x65\x65\x7a\x4a\x2e\xa7\xef\xee\x3e\x7d\xdf\xc5\x98\x1c\x0b\xca\x11\xa6\x8a\xf2\x03\xc3\xa5\x44\x46\x34\x15\x1c\xe5\x14\x96\xd6\x4e\x8c\x59\xc2\x4c\xd4\x1a\xd6\x1b\x48\x7d\x64\xb5\x82\x6b\x51\x56\x94\x21\x68\x5a\x22\x64\x77\x98\xdd\x03\x2d\xc0\x98\xf4\x86\x94\x68\x2d\xd0\xb2\x62\x58\x22\xd7\x0a\x4a\x52\x55\x94\x1f\xd2\xbd\x87\xdf\x75\xe8\x40\xb9\x46\x59\x90\x0c\xd3\xc9\x91\x48\xf8\xf9\xe7\xcc\x0d\x5c\x75\xd0\xc6\x4e\x5c\x03\x9f\x51\xb7\x09\x5f\x45\x8e\xec\x6f\x2b\x16\x35\xcf\x20\x32\x26\xdd\x61\x86\xf4\x88\xd2\x5a\x58\x74\x45\xe2\x11\x7e\xd4\xd2\x02\x8b\x0e\x5e\xcb\x3a\xd3\x9f\x28\xb2\x3c\x86\xa8\x8d\xfa\xec\x04\x50\x4a\x21\x63\x30\x13\x00\x00\x75\xa2\x3a\xbb\x83\x16\x22\xdd\xf2\x1c\x7f\x7d\x7f\xff\xa3\xf9\xd9\x18\x49\xf8\x01\x61\xd6\xd5\x70\x54\xb7\xe5\x95\xe3\xdc\x18\x5a\x00\x17\xfa\x9c\x93\x6e\xd5\x9e\xd1\x0c\xfd\x8b\x38\x94\x8c\x28\x04\x63\x66\xfd\x2a\xd6\xae\x61\xb5\xea\xc5\xc3\x88\xfe\x8e\x3b\x12\x75\x2d\x39\x3c\x5a\xe1\x9b\xf0\x94\xb9\x1a\x57\xc6\x20\xcf\x5d\x27\x4e\x0c\x17\xb4\xa5\x63\xec\x04\x38\x65\xcd\x68\xc8\xf3\x30\x40\xf3\xc1\x87\x73\x2c\x48\xcd\xf4\x7a\xd8\x04\xa7\x2d\x75\x2a\xbd\xc1\x53\x11\x65\x8c\x28\x15\x38\xdd\xf2\x23\x61\x34\xf7\x7c\x27\x30\xad\xa4\x38\xd2\x1c\x73\xa0\x21\xde\xb1\xbb\x86\xf9\x3b\x35\x87\x42\x48\x28\xdd\x3d\xf7\xfd\x76\x3e\x4d\xba\x84\x04\x7a\xcf\x1e\xfb\x26\xec\x24\xe8\x6a\xff\xc6\xba\x1a\xe2\x3f\xad\xab\x24\x8c\x00\x3d\x71\xc5\x81\xa1\x7f\xa6\xad\x26\xef\xcd\x04\x46\x8b\x66\x8a\xcd\xc6\x3d\x70\xd3\x58\x7b\x9e\xd3\xdd\x20\xf9\x79\xf1\xc1\xa6\xd7\xcc\x47\xa2\xf0\xf6\xa1\x0a\x9b\xa3\x8f\x85\x4c\xe1\x6b\x4b\xb4\x02\xbf\x40\xbb\xd0\xf7\x58\xd6\x5d\xd8\x82\x2f\xeb\xb7\xe5\x79\x66\x87\xeb\x5e\x26\xd2\xe2\x8b\x38\xa1\xbc\x26\x25\x32\xe8\xd7\x8d\x83\x8f\x06\x97\x9c\xdb\xc4\xbd\xbb\xeb\x49\x4e\xa3\xc5\xe3\xd3\xc7\x1f\x5c\xda\x90\xfb\x17\xd2\xf9\xf4\x13\x2d\xce\xab\xe1\xb2\xb5\x17\x70\x31\xf4\xff\xd8\xfa\x8d\xeb\x77\x9d\x77\xc7\xc6\x0f\xe2\xd2\x0f\x15\x06\xab\x7b\xeb\x77\x72\x1f\x4f\x34\x6d\x5c\x15\xbf\x66\x43\xfd\xb7\xe5\xe4\xfe\x86\x3d\xcb\xbf\x03\x00\x00\xff\xff\xb8\xbe\xc1\xf6\xb1\x07\x00\x00")

func bindataTemplates08singlerelationertmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates08singlerelationertmpl,
		"templates/08_single-relationer.tmpl",
	)
}



func bindataTemplates08singlerelationertmpl() (*asset, error) {
	bytes, err := bindataTemplates08singlerelationertmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/08_single-relationer.tmpl",
		size: 1969,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588985582, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplates09multirelationertmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x56\x5b\x6b\xdb\x4c\x10\x7d\xf7\xaf\x98\xcf\xe4\xb3\x25\x23\xcb\xe9\xab\xc1\x85\x34\xb4\x25\x90\x4b\x49\x42\x5f\x8c\x09\xaa\x34\x4a\x96\xac\x56\x42\x5a\x39\x0d\xcb\xfe\xf7\xb2\x17\xc9\x96\x6d\xf9\x52\x1c\x48\x4b\xf3\x14\x6b\xb5\x73\x66\xce\x9e\x73\x56\x42\x44\x18\x13\x86\xd0\x4d\x4a\xca\xc9\x30\x47\x1a\x70\x92\x32\xcc\xbb\x30\x94\xb2\x23\xc4\x10\x4e\xd2\x92\xc3\x78\x02\xbe\x7e\x32\x1a\xc1\x79\x9a\x64\x84\x22\x70\x92\x20\x84\x4f\x18\x3e\x43\x9c\xe6\xc0\x9f\x10\x92\x20\xcb\x08\x7b\xf4\xaf\x54\xb5\xdb\xba\x18\x10\xc6\x31\x8f\x83\x10\x81\x24\x19\xc5\x04\x19\xd7\x4b\x7e\x67\x1e\xe4\xf0\xd0\xba\x6f\x02\x3d\x21\xfc\xeb\x20\x41\x29\x85\xec\x28\xf4\xb3\x28\xaa\x5e\xb8\x4a\x23\xa4\x8b\x8a\xc5\x6e\x78\xbf\x13\x97\x2c\x04\x47\x08\xff\x16\x43\x24\x73\xcc\xa5\x84\x41\x8d\xe1\xae\x95\x77\x2a\x4a\x60\x50\x55\xbf\xe3\x79\x19\xf2\x2f\x04\x69\xe4\x41\xa2\x7b\xa8\x81\xd5\x2f\x17\x30\xcf\xd3\x1c\x44\x07\x00\xa0\x78\x21\x3c\x7c\x82\xaa\x8c\x7f\xc1\x22\xfc\x39\x3d\x9d\xd9\x65\x21\xf2\x80\x3d\x22\x9c\xd4\x38\x8a\xea\xaa\x85\x42\x71\x2e\x04\x89\x17\xeb\xfe\x45\x71\x47\x49\x88\xfa\x34\x54\x85\x30\x28\x10\x84\x38\x69\x22\x48\x39\x86\xd1\xa8\xf1\xdc\x8c\xa8\xf7\x18\xe4\x45\x51\xb5\xa4\x80\x1d\x9e\x5e\xa6\x2f\x98\x9f\x07\x09\xd2\x25\xcc\x4f\x41\x81\xf7\xaf\x19\xba\xa6\x9f\xc6\x3e\x29\x3d\x48\x9f\xd5\x6e\xcd\x85\xef\x0c\x96\x41\xab\x9d\x52\xba\x35\x32\x89\xe1\xbf\xf4\xd9\x12\x50\xfd\xe5\xc8\xcb\x9c\x19\xea\x0a\xff\x1a\x5f\x62\x27\xa4\x41\x51\x18\x4e\x35\xdb\x17\x6c\x1e\x50\x12\x7d\x0f\x68\x89\x1e\x74\xb3\x3c\x9d\x93\x08\x23\x20\xe6\x39\xcc\xd5\x02\xf0\xd7\x0c\xc7\xd0\xff\xff\xbe\x0f\xb5\x2e\x63\xb5\x7d\x0c\xfd\x75\x3a\xfa\x5d\x7b\x86\x8b\xf6\x96\x29\x5a\x61\xfe\x5b\xaa\xa5\x54\x73\x6f\xa7\x11\x42\x99\x64\x49\x52\xfe\x3a\x10\x4c\x26\xc0\x08\x5d\x99\x9a\x63\x92\x29\xea\x96\xdf\x37\x74\x09\xd9\x78\x71\x3f\x08\xe8\xa9\x82\x1b\x27\x01\x64\x51\xa3\xef\xd6\xe1\xa4\x1c\x08\x81\x2c\xd2\x27\xbd\x0f\x68\x90\x65\xc8\x22\xa7\xad\xe0\x61\xf5\x3c\x58\xd3\x97\x6b\x9d\x62\x27\xa8\xff\xd1\x8f\x23\x8c\x83\x92\xf2\x71\x67\x3f\x21\x59\x0d\x55\x06\xdb\x24\xa3\x0a\xdc\x8a\xa8\xce\x36\xed\xf4\x7e\x9d\x15\x2b\xca\x31\xdd\x58\x70\x46\x68\xc7\xc4\xd5\x57\xe4\x8d\x3c\x29\x8e\x9c\x57\x6b\xf5\xb7\x07\x96\x0b\x4e\x62\xda\x98\xce\x1a\xa1\xe5\x29\xc2\x0c\x69\xee\xfb\x4f\xae\x95\xb2\x9f\x29\x26\x9b\xac\xa9\x8e\xee\xa1\xca\xe8\xf1\x04\x4c\xc3\x7b\xa9\xba\x69\x53\xcb\x59\x2d\x75\xf3\x7b\x6b\x72\x0c\x01\x69\x81\x6b\xed\x90\x37\xe9\xa3\xe7\xec\x53\x6c\x4a\x66\x6e\x6b\xb3\xca\x9f\xbf\xe3\x33\x46\x8c\x76\xb6\x98\xcd\x5e\x95\xdb\x9d\x56\x18\xa7\xe9\x89\x8c\xf3\xba\x5e\xfd\x82\x0a\x85\xa5\xd9\x36\x38\xae\x62\xa2\xdd\x79\x67\xfc\x8d\xad\x77\xc6\x77\x7d\x2c\x10\xa5\x6f\x05\xb1\xb0\xe1\xdf\x63\xc2\xa3\xdf\x8d\xdb\x05\x56\xcd\xa8\xdb\xbd\x29\xf9\x4d\x7c\xab\x98\xf0\xa0\x6b\x58\x56\x5f\xad\x69\x0c\x59\x5a\x14\xe4\x07\x45\xe3\x39\x1f\xae\xac\xbe\x6c\x5b\x36\xca\x3d\xd0\x27\xb4\x61\xde\xee\x66\xc3\xac\x5e\xa6\x24\xb6\x87\xfb\x11\x28\xb2\x23\x5d\x85\x2e\x0c\xe1\xc3\x1f\x43\x8a\xed\x4c\x8f\xce\x52\xde\x16\xd0\x52\xf6\x6a\x0a\xda\x48\x72\x0e\x61\xa9\xbd\x8c\x6b\xab\x4c\xf5\xf0\x33\x13\x0e\xef\x39\xe2\x56\x86\x3d\x34\xe6\x2e\x91\xbd\x5d\xc4\x5d\x22\xdb\xf9\x69\x41\x18\xf7\xfe\xc5\x57\x53\x3c\xa7\x0b\xd9\xed\x0a\x11\xbb\xe3\x88\x01\xf2\xfe\x25\x7f\x88\xdc\xed\xe4\xbf\x02\x00\x00\xff\xff\xb0\x96\x4f\x5d\xb5\x10\x00\x00")

func bindataTemplates09multirelationertmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplates09multirelationertmpl,
		"templates/09_multi-relationer.tmpl",
	)
}



func bindataTemplates09multirelationertmpl() (*asset, error) {
	bytes, err := bindataTemplates09multirelationertmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/09_multi-relationer.tmpl",
		size: 4277,
		md5checksum: "",
		mode: os.FileMode(436),
		modTime: time.Unix(1588984047, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}


//
// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
//
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
// nolint: deadcode
//
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

//
// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or could not be loaded.
//
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// AssetNames returns the names of the assets.
// nolint: deadcode
//
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

//
// _bindata is a table, holding each asset generator, mapped to its name.
//
var _bindata = map[string]func() (*asset, error){
	"templates/00_imports.tmpl":                bindataTemplates00importstmpl,
	"templates/01_initialize_collections.tmpl": bindataTemplates01initializecollectionstmpl,
	"templates/02_collection.tmpl":             bindataTemplates02collectiontmpl,
	"templates/03_collection-structure.tmpl":   bindataTemplates03collectionstructuretmpl,
	"templates/04_collection-builder.tmpl":     bindataTemplates04collectionbuildertmpl,
	"templates/05_model.tmpl":                  bindataTemplates05modeltmpl,
	"templates/06_primary.tmpl":                bindataTemplates06primarytmpl,
	"templates/07_fielder.tmpl":                bindataTemplates07fieldertmpl,
	"templates/08_single-relationer.tmpl":      bindataTemplates08singlerelationertmpl,
	"templates/09_multi-relationer.tmpl":       bindataTemplates09multirelationertmpl,
}

//
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
//
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, &os.PathError{
					Op: "open",
					Path: name,
					Err: os.ErrNotExist,
				}
			}
		}
	}
	if node.Func != nil {
		return nil, &os.PathError{
			Op: "open",
			Path: name,
			Err: os.ErrNotExist,
		}
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}


type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{Func: nil, Children: map[string]*bintree{
	"templates": {Func: nil, Children: map[string]*bintree{
		"00_imports.tmpl": {Func: bindataTemplates00importstmpl, Children: map[string]*bintree{}},
		"01_initialize_collections.tmpl": {Func: bindataTemplates01initializecollectionstmpl, Children: map[string]*bintree{}},
		"02_collection.tmpl": {Func: bindataTemplates02collectiontmpl, Children: map[string]*bintree{}},
		"03_collection-structure.tmpl": {Func: bindataTemplates03collectionstructuretmpl, Children: map[string]*bintree{}},
		"04_collection-builder.tmpl": {Func: bindataTemplates04collectionbuildertmpl, Children: map[string]*bintree{}},
		"05_model.tmpl": {Func: bindataTemplates05modeltmpl, Children: map[string]*bintree{}},
		"06_primary.tmpl": {Func: bindataTemplates06primarytmpl, Children: map[string]*bintree{}},
		"07_fielder.tmpl": {Func: bindataTemplates07fieldertmpl, Children: map[string]*bintree{}},
		"08_single-relationer.tmpl": {Func: bindataTemplates08singlerelationertmpl, Children: map[string]*bintree{}},
		"09_multi-relationer.tmpl": {Func: bindataTemplates09multirelationertmpl, Children: map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}