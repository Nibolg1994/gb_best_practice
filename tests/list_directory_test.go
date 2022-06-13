package tests

import (
	"fmt"
	"github.com/stretchr/testify"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
	."practic/filesystem"
)

type FileNode struct {
	size  int64
	name  string
	isDir bool
}

type FileEntryNode struct {
	currentNode FileNode
	nodes []*FileNode
}

func (f *FileNode) Name() string {
	return f.name
}

func (f *FileNode) Type() fs.FileMode {
	return 0
}

func (f *FileNode) IsDir() bool {
	return f.isDir
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

func (f *FileNode) Sys() interface{} {
	return nil
}

func (f *FileEntryNode) Info() (fs.FileInfo, error) {
	return &f.currentNode, nil
}

func (f *FileEntryNode) Name() string {
	return f.Name()
}

func (f *FileEntryNode) IsDir() bool {
	return f.IsDir()
}

func (f *FileEntryNode) Type()  fs.FileMode {
	return f.Type()
}

func (f *FileEntryNode) getFile(name string) (result bool, node *FileEntryNode) {
	result = false
	node = nil
	if f.Name() == name {
		return true, f
	}

	for _, file := range f.nodes {
		if file.Name() == name {
			result = true
			node = &FileEntryNode{currentNode: *file, nodes: nil}
			break
		}
	}
	return result, node
}

func (s *FileEntryNode) ReadDir(dirname string) ([]fs.DirEntry, error) {
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

	list := []fs.DirEntry{}

	for _, fileItem := range file.files {
		list = append(list, fileItem)
	}
	return list, nil
}

func TestFileSystem(t *testing.T) {
	root := FileNode{size: 0, name: "root", isDir: true, files: []*FileNode{}}
	varDir := FileNode{size: 0, name: "var", isDir: true, files: []*FileNode{}}
	systemDir := FileNode{size: 0, name: "system", isDir: true, files: []*FileNode{}}
	root.n = append(
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
		ListDirectory(nil, nil, nil, nil, nil),
	)
}