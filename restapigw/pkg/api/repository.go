// Package api -
package api

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
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
)

var (
	parser = config.MakeParser()
)

// ===== [ Types ] =====
type (
	// GroupDefinitions - 리파지토리 Group에 저장된 API Definition 구조 (로드/저장용)
	GroupDefinitions struct {
		Definitions []*config.EndpointConfig `mapstructure:"definitions" yaml:"definitions"`
	}

	// DefinitionMap - 리파지토리의 API Definition 관리 정보 구조 (관리 및 클라이언트 연계용)
	DefinitionMap struct {
		Name        string                   `json:"name"`
		State       ConfigurationState       `json:"-"`
		Definitions []*config.EndpointConfig `json:"definitions"`
	}

	// Repository - Routing 정보 관리 기능을 제공하는 인터페이스 형식
	Repository interface {
		io.Closer

		FindAll() ([]*DefinitionMap, error)
		Write([]*DefinitionMap) error
	}

	// Watcher - Repository에서 API Defintion 변경 여부 감시용
	Watcher interface {
		Watch(ctx context.Context, configurationChan chan<- RepoChangedMessage)
	}

	// Listener - 관리 중인 API Definition 변경 여부 감시용
	Listener interface {
		Listen(ctx context.Context, configurationChan <-chan ConfigChangedMessage)
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

// parseEndpoint - 지정된 정보를 Definition 정보로 전환
func parseEndpoint(apiDef []byte) GroupDefinitions {
	var apiConfigs GroupDefinitions
	log := logging.GetLogger()

	// API 정의들 Unmarshalling
	if err := yaml.Unmarshal(apiDef, &apiConfigs); nil != err {
		// 오류 발생시 단일 Definition으로 다시 처리
		apiConfigs.Definitions = append(apiConfigs.Definitions, config.NewDefinition())
		if err := yaml.Unmarshal(apiDef, &apiConfigs.Definitions[0]); nil != err {
			log.WithError(err).Error("Couldn't parsing api definitions")
		}
	}

	// 로드된 Endpoint 정보 재 구성
	for _, ec := range apiConfigs.Definitions {
		if err := ec.InitializeDefaults(); nil != err {
			log.WithError(err).Error("Couldn't initialize api definition:" + ec.Name)
		}
	}

	return apiConfigs
}

// groupDefinitions - 출력을 위한 Group Definition 구조 반환
func groupDefinitions(dm *DefinitionMap) ([]byte, error) {
	sd := &GroupDefinitions{Definitions: dm.Definitions}
	data, err := yaml.Marshal(sd)
	if nil != err {
		return nil, err
	}
	return data, nil
}

// ===== [ Public Functions ] =====

// BuildRepository - 시스템 설정에 정의된 DSN(Data Group Name) 기준으로 저장소 구성
func BuildRepository(dsn string, refreshTime time.Duration) (Repository, error) {
	log := logging.GetLogger()
	dsnURL, err := url.Parse(dsn)
	if nil != err {
		return nil, errors.Wrap(err, "Error parsing the DSN")
	}
	if "" == dsnURL.Path {
		return nil, errors.New("Path not found from DSN")
	}

	// File 모드인 경우는 상대경로 처리 검증
	if dsnURL.Scheme == "file" && dsnURL.Host == "." {
		path, err := filepath.Abs(dsnURL.Host + dsnURL.Path)
		if nil != err {
			return nil, err
		}
		dsnURL.Path = path
	}

	switch dsnURL.Scheme {
	// CB-STORE (NutsDB or ETCD) 사용
	case cbStore:
		log.Debug("CB-Store (NutsDB or ETCD) based configuration choosen")
		storeKey := dsnURL.Path

		log.WithField("key", storeKey).Debug("Trying to load API configuration files")
		repo, err := NewCbStoreRepository(storeKey, refreshTime)
		if nil != err {
			return nil, errors.Wrap(err, "could not create a CB-Store repository")
		}
		return repo, nil
	// File(Memoery) 사용
	case file:
		log.Debug("File system based configuration choosen")
		apiPath := fmt.Sprintf("%s/apis", dsnURL.Path)

		log.WithField("path", apiPath).Debug("Trying to load API configuration files")
		repo, err := NewFileSystemRepository(apiPath)
		if nil != err {
			return nil, errors.Wrap(err, "could not create a file system repository")
		}
		return repo, nil
	default:
		return nil, errors.New("The selected scheme is not supported to load API definitions")
	}
}
