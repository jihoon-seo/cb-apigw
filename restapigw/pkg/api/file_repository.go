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
	"gopkg.in/yaml.v3"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Endpoint List 구조
	endpointDefinitions struct {
		definitions []*config.EndpointConfig `mapstructure:"definitions" yaml:"definitions"`
	}

	// FileSystemRepository - 파일 시스템 기반 Repository 관리 정보 형식
	FileSystemRepository struct {
		*InMemoryRepository
		watcher *fsnotify.Watcher
	}
)

// ===== [ Implementations ] =====

func (r *FileSystemRepository) parseEndpoint(apiDef []byte) endpointDefinitions {
	apiConfigs := endpointDefinitions{}

	//logging.GetLogger().Printf("YAML: %s", string(apiDef))

	// Try unmarshalling as Array of multiple definitions
	if err := yaml.Unmarshal(apiDef, &apiConfigs); err != nil {
		// Try unmarshalling as Single Definition
		apiConfigs.definitions = append(apiConfigs.definitions, NewEndpoint())
		if err := yaml.Unmarshal(apiDef, &apiConfigs.definitions[0]); err != nil {
			logging.GetLogger().WithError(err).Error("Couldn't parsing api configuration")
		}
	}

	return apiConfigs
}

// Watch - 파일 리파지토리의 대상 파일 변경 감시 및 처리
func (r *FileSystemRepository) Watch(ctx context.Context, configChan chan<- ConfigurationChanged) {
	go func() {
		log := logging.GetLogger()

		for {
			select {
			case event := <-r.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					body, err := ioutil.ReadFile(event.Name)
					if nil != err {
						log.WithError(err).Error("Couldn't load the api defintion file")
						continue
					}
					configChan <- ConfigurationChanged{
						Configurations: &Configuration{Definitions: r.parseEndpoint(body).definitions},
					}
				}
			case err := <-r.watcher.Errors:
				log.WithError(err).Error("error received from file system notify")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Close - 사용 중인 Repository 세션 종료
func (r *FileSystemRepository) Close() error {
	return r.watcher.Close()
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
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a file system watcher")
	}

	repo.watcher = watcher

	for _, f := range files {
		if strings.Contains(f.Name(), ".yaml") {
			filePath := filepath.Join(dir, f.Name())
			logger.WithField("path", filePath)

			appConfigBody, err := ioutil.ReadFile(filePath)
			if err != nil {
				logger.WithError(err).Error("Couldn't load the api definition file")
				return nil, err
			}

			err = repo.watcher.Add(filePath)
			if err != nil {
				logger.WithError(err).Error("Couldn't load the api definition file")
				return nil, err
			}

			definition := repo.parseEndpoint(appConfigBody)
			for _, v := range definition.definitions {
				if err = repo.add(v); err != nil {
					logger.WithField("endpoint", v.Endpoint).WithError(err).Error("Failed during add endpoint to the repository")
					return nil, err
				}
			}
		}
	}

	return &repo, nil
}
