// Package api -
package api

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"gopkg.in/fsnotify.v1"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (

	// FileSystemRepository - 파일 시스템 기반 Repository 관리 정보 형식
	FileSystemRepository struct {
		*InMemoryRepository
		watcher    *fsnotify.Watcher
		sourcePath string
	}
)

// ===== [ Implementations ] =====

// Write - 변경된 리파지토리 내용을 대상 파일로 출력
func (fsr *FileSystemRepository) Write(definitionMaps []*DefinitionMap) error {
	fsr.Sources = definitionMaps

	for _, dm := range fsr.Sources {
		if dm.State == "REMOVED" {
			err := os.Remove(path.Join(fsr.sourcePath, dm.Source))
			if nil != err {
				return err
			}
		} else if dm.State != "" {
			data, err := sourceDefinitions(dm)
			if nil != err {
				return err
			}
			err = ioutil.WriteFile(path.Join(fsr.sourcePath, dm.Source), data, os.FileMode(0666))
			if nil != err {
				return err
			}
		}
	}

	return nil
}

// Watch - 파일 리파지토리의 대상 파일 변경 감시 및 처리
func (fsr *FileSystemRepository) Watch(ctx context.Context, configChan chan<- ConfigurationChanged) {
	go func() {
		log := logging.GetLogger()

		for {
			select {
			case event := <-fsr.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					body, err := ioutil.ReadFile(event.Name)
					if nil != err {
						log.WithError(err).Error("Couldn't load the api defintion file")
						continue
					}
					configChan <- ConfigurationChanged{
						Configurations: &Configuration{DefinitionMaps: []*DefinitionMap{
							{
								Source:      event.Name,
								State:       "",
								Definitions: parseEndpoint(body).Definitions,
							},
						}},
					}
				}
			case err := <-fsr.watcher.Errors:
				log.WithError(err).Error("error received from file system notify")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Close - 사용 중인 Repository 세션 종료
func (fsr *FileSystemRepository) Close() error {
	return fsr.watcher.Close()
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewFileSystemRepository - 파일 시스템 기반의 Repository 인스턴스 생성
func NewFileSystemRepository(dir string) (*FileSystemRepository, error) {
	logger := logging.GetLogger()
	repo := FileSystemRepository{InMemoryRepository: NewInMemoryRepository(), sourcePath: dir}

	// Grab json files from directory
	files, err := ioutil.ReadDir(dir)
	if nil != err {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if nil != err {
		return nil, errors.Wrap(err, "failed to create a file system watcher")
	}

	repo.watcher = watcher

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") {
			fileName := f.Name()
			filePath := filepath.Join(dir, fileName)
			logger.WithField("path", filePath)

			appConfigBody, err := ioutil.ReadFile(filePath)
			if nil != err {
				logger.WithError(err).Error("Couldn't load the api definition file")
				return nil, err
			}

			err = repo.watcher.Add(filePath)
			if nil != err {
				logger.WithError(err).Error("Couldn't load the api definition file")
				return nil, err
			}

			definition := parseEndpoint(appConfigBody)
			for _, v := range definition.Definitions {
				if err = repo.add(fileName, v); nil != err {
					logger.WithField("endpoint", v.Endpoint).WithError(err).Error("Failed during add endpoint to the repository")
					return nil, err
				}
			}
		}
	}

	return &repo, nil
}
