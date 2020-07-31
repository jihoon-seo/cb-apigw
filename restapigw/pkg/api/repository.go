// Package api -
package api

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====

const (
	file    = "file"
	cbStore = "cbstore"
	db      = "database"
)

var (
	parser = config.MakeParser()
)

// ===== [ Types ] =====
type (
	// DefinitionMap - 리파지토리의 API Definition 관리 정보 구조
	DefinitionMap struct {
		Source      string                   `json:"source"`
		HasChanges  bool                     `json:"-"`
		Definitions []*config.EndpointConfig `json:"definitions"`
	}

	// Repository - Routing 정보 관리 기능을 제공하는 인터페이스 형식
	Repository interface {
		io.Closer

		//FindSources() ([]string, error)
		FindAll() ([]*DefinitionMap, error)
		//FindAllBySource(source string) ([]*config.EndpointConfig, error)
		Write() error
	}

	// Watcher - Routing 정보 변경을 감시하는 기능을 제공하는 인터페이스 형식
	Watcher interface {
		Watch(ctx context.Context, configurationChan chan<- ConfigurationChanged)
	}

	// Listener - Configuration 변경을 처리하기 위한 Listaner 인터페이스 형식
	Listener interface {
		Listen(ctx context.Context, configurationChan <-chan ChangeMessage)
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// BuildRepository - 시스템 설정에 정의된 DSN(Data Source Name) 기준으로 저장소 구성
func BuildRepository(dsn string, refreshTime time.Duration) (Repository, error) {
	logger := logging.GetLogger()
	dsnURL, err := url.Parse(dsn)
	if nil != err {
		return nil, errors.Wrap(err, "Error parsing the DSN")
	}

	switch dsnURL.Scheme {
	// CB-STORE (NutsDB or ETCD) 사용
	case cbStore:
		logger.Debug("CB-Store (NutsDB or ETCD) based configuration choosen")
		apiKey := dsnURL.Path

		logger.WithField("key", apiKey).Debug("Trying to load API configuration files")
		repo, err := NewCbStoreRepository(apiKey)
		if nil != err {
			return nil, errors.Wrap(err, "could not create a CB-Store repository")
		}
		return repo, nil
	// File(Memoery) 사용
	case file:
		logger.Debug("File system based configuration choosen")
		apiPath := fmt.Sprintf("%s/apis", dsnURL.Path)

		logger.WithField("path", apiPath).Debug("Trying to load API configuration files")
		repo, err := NewFileSystemRepository(apiPath)
		if nil != err {
			return nil, errors.Wrap(err, "could not create a file system repository")
		}
		return repo, nil
	// Database 사용
	case db:
		logger.Debug("File system based configuration choosen")
		return nil, errors.New("The database scheme is not supported to load API definitions")
	default:
		return nil, errors.New("The selected scheme is not supported to load API definitions")
	}
}
