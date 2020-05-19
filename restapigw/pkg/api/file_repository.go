// Package api -
package api

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"gopkg.in/fsnotify.v1"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Type used for JSON.Unmarshaller
	definitionList struct {
		defs []*Definition
	}

	// FileSystemRepository - 파일 시스템 기반 Repository 관리 정보 형식
	FileSystemRepository struct {
		*InMemoryRepository
		watcher *fsnotify.Watcher
	}
)

// ===== [ Implementations ] =====

func (r *FileSystemRepository) parseDefinition(apiDef []byte) definitionList {
	appConfigs := definitionList{}

	// Try unmarshalling as if json is an unnamed Array of multiple definitions
	if err := json.Unmarshal(apiDef, &appConfigs); err != nil {
		// Try unmarshalling as if json is a single Definition
		appConfigs.defs = append(appConfigs.defs, NewDefinition())
		if err := json.Unmarshal(apiDef, &appConfigs.defs[0]); err != nil {
			logging.GetLogger().WithError(err).Error("[RPC] --> Couldn't unmarshal api configuration")
		}
	}

	return appConfigs
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
		if strings.Contains(f.Name(), ".json") {
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

			definition := repo.parseDefinition(appConfigBody)
			for _, v := range definition.defs {
				if err = repo.add(v); err != nil {
					logger.WithField("name", v.Name).WithError(err).Error("Failed during add definition to the repository")
					return nil, err
				}
			}
		}
	}

	return &repo, nil
}
