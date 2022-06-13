package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"path/filepath"
	. "practic/filesystem"
	"runtime"
	"syscall"
	"time"
)

func GetLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"debug.log"}
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, _ := config.Build()
	logger = logger.With(zap.String("goos", runtime.GOOS))
	return logger
}

func FindFiles(logger *zap.Logger, cancelCtx context.Context, userChan chan struct{}, ext string) (FileList, error) {
	wd, err := os.Getwd()
	logger.With(
		zap.String("time", time.Now().String()),
		zap.String("dir", wd),
	).Debug("FindFiles, call")

	if err != nil {
		return nil, err
	}
	files, err := ListDirectory(logger, cancelCtx, userChan, wd, 1)
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
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.With(
				zap.String("time", time.Now().String()),
			).Warn("Logger is not sync" + err.Error())
		}
	}(logger)

	const wantExt = ".go"
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	sigUserCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigUserCh, syscall.SIGTERM)

	//Обработать сигнал SIGUSR1
	waitCh := make(chan struct{})
	userCh := make(chan struct{})
	go func() {
		res, err := FindFiles(logger, ctx, userCh, wantExt)
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
			logger.With(
				zap.String("time", time.Now().String()),
			).Debug("Signal received, terminate...")
			cancel()
		case <-sigUserCh:
			logger.With(
				zap.String("time", time.Now().String()),
			).Debug("Received user signal...")
			userCh <- struct{}{}
		}
	}()
	//Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	logger.With(
		zap.String("time", time.Now().String()),
	).Debug("Done")
}
