package main

import (
	"fmt"
	testify "github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
)


type FileNode struct {
	size  int64
	name  string
	isDir bool
	files []*FileNode
}

func (f *FileNode) Name() string {
	return f.name
}

func (f *FileNode) Size() int64 {
	return f.size
}

func (f *FileNode) Mode() fs.FileMode {
	return 0
}

func (f *FileNode) ModTime() time.Time {
	return time.Now()
}

func (f *FileNode) IsDir() bool {
	return f.isDir
}

func (f *FileNode) Sys() interface{} {
	return nil
}

func (f *FileNode) getFile(name string) (result bool, node *FileNode) {
	result = false
	node = nil
	if f.name == name {
		return true, f
	}

	for _, file := range f.files {
		if file.name == name {
			result = true
			node = file
			break
		}
	}
	return result, node
}

func (s *FileNode) ReadDir(dirname string) ([]fs.FileInfo, error) {
	items := strings.Split(dirname, string(os.PathSeparator))
	file := s
	ok := false
	for _, item := range items {
		if item == "" {
			continue
		}
		ok, file = file.getFile(item)
		if !ok {
			return nil, fmt.Errorf("file %s not found ", dirname)
		}
		if !file.IsDir() {
			return nil, fmt.Errorf("file %s is not dir ", dirname)
		}
	}

	list := []fs.FileInfo{}

	for _, fileItem := range file.files {
		list = append(list, fileItem)
	}
	return list, nil
}

func TestFileSystem(t *testing.T) {
	root := FileNode{size: 0, name: "root", isDir: true, files: []*FileNode{}}
	varDir := FileNode{size: 0, name: "var", isDir: true, files: []*FileNode{}}
	systemDir := FileNode{size: 0, name: "system", isDir: true, files: []*FileNode{}}
	root.files = append(
		root.files,
		&varDir,
		&systemDir,
		&FileNode{size: 0, name: "a.txt", isDir: false, files: []*FileNode{}},
		&FileNode{size: 0, name: "b.txt", isDir: false, files: []*FileNode{}},
	)

	varDir.files = append(
		varDir.files,
		&FileNode{size: 0, name: "a.txt", isDir: false, files: []*FileNode{}},
		&FileNode{size: 0, name: "b.txt", isDir: false, files: []*FileNode{}},
	)

	systemDir.files = append(
		systemDir.files,
		&FileNode{size: 0, name: "a.txt", isDir: false, files: []*FileNode{}},
		&FileNode{size: 0, name: "b.txt", isDir: false, files: []*FileNode{}},
	)

	_, err := root.ReadDir("/root/var/")

	if err != nil {
		t.Fatalf("can not read files")
	}

	_, err = root.ReadDir("/root/var")
	if err != nil {
		t.Fatalf("can not read files")
	}

	_, err = root.ReadDir("/root/var/a.txt")
	if err == nil {
		t.Fatalf("can not read file as dir")
	}

}

func TestListDirectory(t *testing.T) {
	root := &FileNode{size: 0, name: "root", isDir: true, files: []*FileNode{}}
	varDir := FileNode{size: 0, name: "var", isDir: true, files: []*FileNode{}}
	systemDir := FileNode{size: 0, name: "system", isDir: true, files: []*FileNode{}}
	root.files = append(root.files,
		&varDir,
		&systemDir,
		&FileNode{size: 0, name: "a.txt", isDir: false, files: []*FileNode{}},
		&FileNode{size: 0, name: "b.txt", isDir: false, files: []*FileNode{}},
	)

	varDir.files = append(
		varDir.files,
		&FileNode{size: 0, name: "a.txt", isDir: false, files: []*FileNode{}},
		&FileNode{size: 0, name: "b.txt", isDir: false, files: []*FileNode{}},
	)

	systemDir.files = append(
		systemDir.files,
		&FileNode{size: 0, name: "a.txt", isDir: false, files: []*FileNode{}},
		&FileNode{size: 0, name: "b.txt", isDir: false, files: []*FileNode{}},
	)

	testify.Equal(
		t,
		[]string{"/root/system/a.txt", "/root/system/b.txt", "/root/var/a.txt", "/root/var/b.txt"},
		ListDirectory(),
	)
}