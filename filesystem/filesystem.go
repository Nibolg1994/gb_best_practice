package filesystem

import (
	"context"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

type TargetFile struct {
	Path string
	Name string
}

type FileList map[string]TargetFile

type FileInfo interface {
	os.FileInfo
	Path() string
}

type fileInfo struct {
	os.FileInfo
	path string
}

func (fi fileInfo) Path() string {
	return fi.path
}

func ListDirectory(logger *zap.Logger, cancelCtx context.Context, userChan chan struct{}, dir string, depth int) ([]FileInfo, error) {

	if logger != nil {
		logger.With(
			zap.String("time", time.Now().String()),
			zap.String("dir", dir),
		).Debug("ListDirectory, call")
	}

	select {
	case <-cancelCtx.Done():
		if logger != nil {
			logger.With(
				zap.String("time", time.Now().String()),
			).Debug("Context closed")
		}
		return nil, nil
	case <-userChan:
		if logger != nil {
			logger.With(
				zap.String("time", time.Now().String()),
				zap.String("dir", dir),
				zap.String("depth", dir),
			).Debug("processing User signal")
		}
	}

	time.Sleep(time.Second * 2)
	var result []FileInfo
	res, err := os.ReadDir(dir)
	if logger != nil {
		logger.With(
			zap.String("time", time.Now().String()),
		).Debug("Found file ...")
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range res {
		path := filepath.Join(dir, entry.Name())
		if logger != nil {
			logger.With(
				zap.String("time", time.Now().String()),
				zap.String("path", path),
			).Debug("Found file ...")
		}
		if entry.IsDir() {
			depth++
			child, err := ListDirectory(logger, cancelCtx, userChan, path, depth) //Дополнительно: вынести в горутину
			if err != nil {
				return nil, err
			}
			result = append(result, child...)
		} else {
			info, _ := entry.Info()
			result = append(result, fileInfo{info, path})
		}
	}
	return result, nil
}
