package file

import (
	"io/fs"
	"time"
)

type File struct {
	Name        string
	CreatedTime time.Time
}

type Files []*File

func NewFile(file fs.FileInfo) *File {
	return &File{
		Name:        file.Name(),
		CreatedTime: file.ModTime(),
	}
}
