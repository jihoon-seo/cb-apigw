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
	"gopkg.in/yaml.v2"
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
	// SourceDefinitions - 리파지토리 Source에 저장된 API Definition 구조 (로드/저장용)
	SourceDefinitions struct {
		Definitions []*config.EndpointConfig `mapstructure:"definitions" yaml:"definitions"`
	}

	// DefinitionMap - 리파지토리의 API Definition 관리 정보 구조 (관리 및 클라이언트 연계용)
	DefinitionMap struct {
		Source      string                   `json:"source"`
		State       string                   `json:"-"`
		Definitions []*config.EndpointConfig `json:"definitions"`
	}

	// Repository - Routing 정보 관리 기능을 제공하는 인터페이스 형식
	Repository interface {
		io.Closer

		FindAll() ([]*DefinitionMap, error)
		Write([]*DefinitionMap) error
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

// parseEndpoint - 지정된 정보를 Definition 정보로 전환
func parseEndpoint(apiDef []byte) SourceDefinitions {
	var apiConfigs SourceDefinitions
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

// sourceDefinitions - 출력을 위한 Source Definition 구조 반환
func sourceDefinitions(dm *DefinitionMap) ([]byte, error) {
	sd := &SourceDefinitions{Definitions: dm.Definitions}
	data, err := yaml.Marshal(sd)
	if nil != err {
		return nil, err
	}
	return data, nil
}

// ===== [ Public Functions ] =====

// BuildRepository - 시스템 설정에 정의된 DSN(Data Source Name) 기준으로 저장소 구성
func BuildRepository(dsn string, refreshTime time.Duration) (Repository, error) {
	logger := logging.GetLogger()
	dsnURL, err := url.Parse(dsn)
	if nil != err {
		return nil, errors.Wrap(err, "Error parsing the DSN")
	}

	if "" == dsnURL.Path {
		return nil, errors.New("Path not found from DSN")
	}

	switch dsnURL.Scheme {
	// CB-STORE (NutsDB or ETCD) 사용
	case cbStore:
		logger.Debug("CB-Store (NutsDB or ETCD) based configuration choosen")
		storeKey := dsnURL.Path

		logger.WithField("key", storeKey).Debug("Trying to load API configuration files")
		repo, err := NewCbStoreRepository(storeKey)
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
