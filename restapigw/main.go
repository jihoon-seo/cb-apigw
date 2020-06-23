package main

import (
	"log"
	"os"

	"github.com/cloud-barista/cb-apigw/restapigw/cmd"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====
func init() {
	// FIXME: CBLOG 경로관련 문제 (현재 경로로 환경변수 설정)
	if dir, err := os.Getwd(); err == nil {
		os.Setenv("CBLOG_ROOT", dir)
		os.Setenv("CBSTORE_ROOT", dir)
	}
}

// main - Entrypoint
func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Println("cb-restapigw terminated with error: ", err.Error())
	}
}

// ===== [ Public Functions ] =====
