package main

import (
	"bytes"
	"io/ioutil"
	"strings"
)

type InputStream interface {
	Next() *bytes.Buffer
}

// NovelInputStream
// "Novel" in the sense of a book
// Type of input stream which writes the contents of a book one sentence at a time
type FileInputStream struct {
	filename string
	text     []string
	ptr      int
}

func NewNovelInputStream(file string) (*FileInputStream, error) {
	// Even a long book will only be a few Mb as a txt file, so we can safely
	// assume that the file will safely fit into memory
	book, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	text := strings.Split(string(book), ".")

	return &FileInputStream{
		filename: file,
		text:     text,
		ptr:      0,
	}, nil
}

func (s *FileInputStream) Next() *bytes.Buffer {
	buf := bytes.NewBufferString(s.text[s.ptr])
	s.ptr += 1
	if s.ptr >= len(s.text) {
		s.ptr = 0
	}
	return buf
}

// NilInputStream
// This type should be used for requests without a request body
type NilInputStream struct{}

func (s *NilInputStream) Next() *bytes.Buffer {
	return nil
}

func GetInputStream(source string, file string) (InputStream, error) {
	switch source {
	case "novel":
		return NewNovelInputStream(file)
	default:
		return &NilInputStream{}, nil
	}
}
