// Package api -
package api

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v2"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Endpoint List 구조
	endpointDefinitions struct {
		Definitions []*config.EndpointConfig `mapstructure:"definitions" yaml:"definitions"`
	}

	// FileSystemRepository - 파일 시스템 기반 Repository 관리 정보 형식
	FileSystemRepository struct {
		*InMemoryRepository
		watcher *fsnotify.Watcher
	}
)

// ===== [ Implementations ] =====

func (fsr *FileSystemRepository) parseEndpoint(apiDef []byte) endpointDefinitions {
	var apiConfigs endpointDefinitions
	log := logging.GetLogger()

	// API 정의들 Unmarshalling
	if err := yaml.Unmarshal(apiDef, &apiConfigs); nil != err {
		// 오류 발생시 단일 Definition으로 다시 처리
		apiConfigs.Definitions = append(apiConfigs.Definitions, NewDefinition())
		if err := yaml.Unmarshal(apiDef, &apiConfigs.Definitions[0]); nil != err {
			log.WithError(err).Error("Couldn't parsing api definitions")
		}
	}

	// 로드된 Endpoint 정보 출력
	for idx, ec := range apiConfigs.Definitions {
		log.Debugf("Endpoint(%d) : %+v\n", idx, ec)
		for bIdx, bc := range ec.Backend {
			log.Debugf("Backend(%d) : %v\n", bIdx, bc)
		}
	}

	return apiConfigs
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
						Configurations: &Configuration{Definitions: fsr.parseEndpoint(body).Definitions},
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
	repo := FileSystemRepository{InMemoryRepository: NewInMemoryRepository()}

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
		if strings.Contains(f.Name(), ".yaml") {
			filePath := filepath.Join(dir, f.Name())
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

			definition := repo.parseEndpoint(appConfigBody)
			for _, v := range definition.Definitions {
				if err = repo.add(v); nil != err {
					logger.WithField("endpoint", v.Endpoint).WithError(err).Error("Failed during add endpoint to the repository")
					return nil, err
				}
			}
		}
	}

	return &repo, nil
}
