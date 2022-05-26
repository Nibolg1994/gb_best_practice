package main

//Исходники задания для первого занятия у других групп https://github.com/t0pep0/GB_best_go

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func GetLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := config.Build()
	logger = logger.With(zap.String("goos", runtime.GOOS))
	return logger
}

//Ограничить глубину поиска заданым числом, по SIGUSR2 увеличить глубину поиска на +2
func ListDirectory(cancelCtx context.Context, userChan chan struct{}, dir string, depth int) ([]FileInfo, error) {
	logger := GetLogger()
	defer logger.Sync()

	logger.With(
		zap.String("time", time.Now().String()),
		zap.String("dir", dir),
	).Debug("ListDirectory, call")

	select {
	case <-cancelCtx.Done():
		return nil, nil
	case <-userChan:
		fmt.Println(dir)
		fmt.Println(depth)
	}

	//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
	time.Sleep(time.Second * 10)
	var result []FileInfo
	res, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range res {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			depth++
			child, err := ListDirectory(cancelCtx, userChan, path, depth) //Дополнительно: вынести в горутину
			if err != nil {
				return nil, err
			}
			result = append(result, child...)
		} else {
			result = append(result, fileInfo{entry, path})
		}
	}
	return result, nil
}

func FindFiles(cancelCtx context.Context, userChan chan struct{}, ext string) (FileList, error) {
	logger := GetLogger()
	defer logger.Sync()

	wd, err := os.Getwd()
	logger.With(
		zap.String("time", time.Now().String()),
		zap.String("dir", wd),
	).Debug("FindFiles, call")

	if err != nil {
		return nil, err
	}
	files, err := ListDirectory(cancelCtx, userChan, wd, 1)
	if err != nil {
		return nil, err
	}
	fl := make(FileList, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			fl[file.Name()] = TargetFile{
				Name: file.Name(),
				Path: file.Path(),
			}
		}
	}
	return fl, nil
}

func main() {
	logger := GetLogger()
	defer logger.Sync()

	const wantExt = ".go"
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	sigUserCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigUserCh, syscall.SIGUSR1)

	//Обработать сигнал SIGUSR1
	waitCh := make(chan struct{})
	userCh := make(chan struct{})
	go func() {
		res, err := FindFiles(ctx, userCh, wantExt)
		if err != nil {
			logger.Fatal(err.Error())
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
		}
		waitCh <- struct{}{}
	}()
	go func() {
		select {
		case <-sigCh:
			log.Println("Signal received, terminate...")
			cancel()
		case <-sigUserCh:
			userCh <- struct{}{}
		}
	}()
	//Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	log.Println("Done")
}
