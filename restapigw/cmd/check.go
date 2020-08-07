package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"

	"github.com/spf13/cobra"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// cheeckAndLoad - 지정된 Command 와 args 를 기준으로 Configuration 파일 검증 및 로드
func checkAndLoad(cmd *cobra.Command, args []string) (config.ServiceConfig, error) {
	var (
		sConf config.ServiceConfig
		err   error
	)

	if configFile == "" {
		cmd.Println("[CHECK - ERROR] Please, provide the path to your configuration file")
		return sConf, errors.New("configuration file are not specified")
	}

	cmd.Printf("[CHECK] Parsing configuration file: %s\n", configFile)
	if sConf, err = parser.Parse(configFile); err != nil {
		cmd.Println("[CHECK - ERROR] Parsing the configuration file.\n", err.Error())
		return sConf, err
	}

	// Command line 에 지정된 '-d', '-p' 옵션을 설정에 적용 (우선권)
	sConf.Debug = sConf.Debug || debug
	if port != 0 {
		sConf.Port = port
	}

	// TODO: Check Items and Prints
	//err = checkAndPrintServiceConf(cmd, sConf)

	return sConf, err
}

// checkAndPrintBackendConf - Backend 설정 정보를 검증하고 출력
func checkAndPrintBackendConf(cmd *cobra.Command, eConf *config.EndpointConfig) error {
	if 0 == len(eConf.Backend) {
		return errors.New("[CHECK - ERROR] No backend configuration, must be at least one backend")
	}

	cmd.Printf("\t\tBackends (%d):\n", len(eConf.Backend))
	for _, backend := range eConf.Backend {
		if 0 == len(backend.Hosts) {
			return errors.New("No backend host or global service host in configuration, must be specified host in backend or service")
		}

		cmd.Printf("\t\t\tHosts: %v, URL: %s, Method: %s\n", backend.Hosts, backend.URLPattern, backend.Method)
		cmd.Printf("\t\t\t\tTimeout: %s, Target: %s, Mapping: %v, Blacklist: %v, Whitelist: %v, Group: %v\n",
			backend.Timeout, backend.Target, backend.Mapping, backend.Blacklist, backend.Whitelist,
			backend.Group)

		cmd.Printf("\t\t\tMiddleware (%d):\n", len(backend.Middleware))
		for k, v := range backend.Middleware {
			cmd.Printf("\t\t\t\t%s: %v\n", k, v)
		}
	}

	return nil
}

// checkFunc - 지정된 args 에서 설정과 관련된 정보를 로드/검증/출력 처리
func checkFunc(cmd *cobra.Command, args []string) {
	var (
		err error
	)

	if _, err = checkAndLoad(cmd, args); err != nil {
		fmt.Printf("[CHECK - ERROR] %s \n", err)
		os.Exit(1)
		return
	}

	cmd.Println("[CHECK] Syntax OK!")
}

// ===== [ Public Functions ] =====

// NewCheckCmd - 설정 검증 기능을 수행하는 Cobra Command 생성
func NewCheckCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "check",
		Short:   "Validates that the configuration file is valid",
		Long:    "Validates that the active configuration file has  a valid syntax to run the service. \nChange the configuration file by using the --config flag (default $PWD/conf/cb-restapigw.yaml)",
		Run:     checkFunc,
		Aliases: []string{"validate"},
		Example: core.AppName + " check --config config.yaml",
	}
}
