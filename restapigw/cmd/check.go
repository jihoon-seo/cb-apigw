package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"

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
		err  error
	)

	if configFile == "" {
		cmd.Println("Please, provide the path to your configuration file")
		return sConf, errors.New("configuration file are not specified")
	}

	cmd.Printf("Parsing configuration file: %s\n", configFile)
	if sConf, err = parser.Parse(configFile); err != nil {
		cmd.Println("ERROR - Parsing the configuration file.\n", err.Error())
		return sConf, err
	}

	// Command line 에 지정된 '-d', '-p' 옵션을 설정에 적용 (우선권)
	sConf.Debug = sConf.Debug || debug
	if port != 0 {
		sConf.Port = port
	}

	return sConf, nil
}

// printConfigurations - 지정된 설정 정보를 Debug 용도로 출력
func printConfigurations(cmd *cobra.Command, sConf config.ServiceConfig) {
	cmd.Printf("Parsed configuration: CacheTTL: %s, Port: %d\n", sConf.CacheTTL.String(), sConf.Port)
	cmd.Printf("Hosts: %v\n", sConf.Host)

	cmd.Printf("Moddleware (%d):\n", len(sConf.Middleware))
	for k, v := range sConf.Middleware {
		cmd.Printf("\t%s: %v\n", k, v)
	}

	cmd.Printf("Endpoints (%d):\n", len(sConf.Endpoints))
	for _, endpoint := range sConf.Endpoints {
		cmd.Printf("\tEndpoint: %s, Method: %s, CacheTTL: %s, Querystring: %v\n",
			endpoint.Endpoint, endpoint.Method, endpoint.CacheTTL.String(), endpoint.QueryString)

		cmd.Printf("\t\tModdleware (%d):\n", len(endpoint.Middleware))
		for k, v := range endpoint.Middleware {
			cmd.Printf("\t\t\t%s: %v\n", k, v)
		}

		cmd.Printf("\t\tBackends (%d):\n", len(endpoint.Backend))
		for _, backend := range endpoint.Backend {
			cmd.Printf("\t\t\tHosts: %v, URL: %s, Method: %s\n", backend.Host, backend.URLPattern, backend.Method)
			cmd.Printf("\t\t\t\tTimeout: %s, Target: %s, Mapping: %v, Blacklist: %v, Whitelist: %v, Group: %v\n",
				backend.Timeout, backend.Target, backend.Mapping, backend.Blacklist, backend.Whitelist,
				backend.Group)

			cmd.Printf("\t\t\tMiddleware (%d):\n", len(backend.Middleware))
			for k, v := range backend.Middleware {
				cmd.Printf("\t\t\t\t%s: %v\n", k, v)
			}
		}
	}
}

// checkFunc - 지정된 args 에서 설정과 관련된 정보를 로드/검증/출력 처리
func checkFunc(cmd *cobra.Command, args []string) {
	var (
		sConf config.ServiceConfig
		err  error
	)

	if sConf, err = checkAndLoad(cmd, args); err != nil {
		os.Exit(1)
		return
	}

	if debug {
		printConfigurations(cmd, sConf)
	}

	cmd.Println("Syntax OK!")
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
