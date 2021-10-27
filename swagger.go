package grok

import (
	"bytes"
	"io"
	"os"
)

var (
	swagger string
)

//SwaggerDoc ...
type SwaggerDoc struct {
	Path string
}

//NewSwaggerDoc ...
func NewSwaggerDoc(path string) *SwaggerDoc {
	return &SwaggerDoc{
		Path: path,
	}
}

//ReadDoc ...
func (s *SwaggerDoc) ReadDoc() string {
	if swagger != "" {
		return swagger
	}

	buf := bytes.NewBuffer(nil)

	f, err := os.Open(s.Path)

	if err != nil {
		panic(err)
	}

	io.Copy(buf, f)
	f.Close()

	swagger = string(buf.Bytes())

	return swagger
}
