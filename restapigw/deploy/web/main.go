package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// ===== [ Constants and Variables ] =====

var (
	task      config
	secretKey string
)

// ===== [ Types ] =====

// config - HMAC 운영을 위한 데이터 처리 구조
type config struct {
	SecretKey string `mapstructure:"secret_key" query:"SecretKey"`
	AccessKey string `mapstructure:"access_key" query:"AccessKey"`
	Duration  string `mapstructure:"duration" query:"Duration"`
	Timestamp string `mapstructure:"timestamp" query:"Timestamp"`
	Token     string `mapstructure:"token" query:"Token"`
	Message   string `mapstructure:"message" query:"Message"`
}

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// getTime - UTC 기준 현재 시간 반환
func getTime() time.Time {
	return time.Now().UTC()
}

// parseDuration - 지정한 Time 제한을 time.Duration 값으로 반환
func parseDuration(limitTime string) time.Duration {
	duration, err := time.ParseDuration(limitTime)
	if nil != err {
		return 0
	}

	return duration
}

// checkDuration - 지정한 시간과 Duration의 초과여부 검증
func checkDuration(checkTime time.Time, timestamp string, duration time.Duration) bool {
	ts, err := time.Parse(time.UnixDate, timestamp)
	if nil != err {
		return false
	}

	ts = ts.Add(duration)

	log.Printf("current Time: %v", checkTime)
	log.Printf("Durable timestamp: %v", ts)
	log.Printf("Difference: %v", checkTime.Sub(ts))

	return 0 <= ts.Sub(checkTime)
}

// makeToken - 현재 시간 + 제한 시간 + 액세스 키를 기준으로 비밀 키를 사용해서 HMAC 토큰 생성
func makeToken(task config) []byte {
	data := task.Duration + "^" + task.Timestamp + "^" + task.AccessKey

	// Create HMAC
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))

	return h.Sum(nil)
}

// getHMACToken - 생성한 토큰을 검증 정보 (현재 시간, 제한 시간, 액세스 키)를 추가한 최종 HMAC 토큰 문자열 반환
func getHMACToken(task config) config {
	token := append(makeToken(task), []byte("||"+task.Timestamp+"|"+task.Duration+"|"+task.AccessKey)...)
	task.Token = hex.EncodeToString(token)

	return task
}

// getTokenData - 지정한 토큰 데이터를 검증 가능한 [][]byte로 반환
func getTokenData(token string) [][]byte {
	log.Printf("received token: [%v]", token)

	tokenBytes, err := hex.DecodeString(token)
	if nil != err {
		return [][]byte{}
	}

	sep := []byte("||")

	data := bytes.Split(tokenBytes, sep)

	return data
}

// getTask - 설정 정보 반환
func getConfigInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, task)
	}
}

// createToken - 요청된 정보를 기준으로 HMAC 토큰 문자열 생성
func createToken() echo.HandlerFunc {
	return func(c echo.Context) error {
		var newTask config
		c.Bind(&newTask)

		newTask.Timestamp = getTime().Format(time.UnixDate)

		return c.JSON(http.StatusOK, getHMACToken(newTask))
	}
}

// validateToken - 요청된 정보를 기준으로 HMAC 토큰 문자열 유효성 검증
func validateToken() echo.HandlerFunc {
	return func(c echo.Context) error {
		var newTask config
		c.Bind(&newTask)

		var tokenData = getTokenData(newTask.Token)
		if len(tokenData[0]) == 0 {
			newTask.Message = "Token data not founded."
		}

		tokenInfo := strings.Split(string(tokenData[1]), "|")
		currTime := getTime()

		newTask.Timestamp = tokenInfo[0]
		newTask.Duration = tokenInfo[1]
		newTask.AccessKey = tokenInfo[2]
		newToken := makeToken(newTask)
		if !bytes.Equal(newToken, tokenData[0]) {
			newTask.Message = "Invalid token."
		} else if !checkDuration(currTime, newTask.Timestamp, parseDuration(newTask.Duration)) {
			newTask.Message = "Time limit excceeded."
		}

		return c.JSON(http.StatusOK, newTask)
	}
}

// loadConfig - 구동을 위한 설정 정보 로드
func loadConfig() {
	viper.SetConfigFile("./conf/hmac.yaml")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")

	task = config{}

	// Reading
	if err := viper.ReadInConfig(); nil != err {
		task = config{}
	}
	// Unmarshal to struct
	if err := viper.Unmarshal(&task); nil != err {
		task = config{}
	}

	secretKey = task.SecretKey
	task.SecretKey = ""
}

// main - Entry point
func main() {
	loadConfig()

	e := echo.New()

	e.File("/", "public/index.html")

	e.GET("/task", getConfigInfo())
	e.PUT("/task", createToken())
	e.GET("/validate", validateToken())

	e.Logger.Fatal(e.Start(":8010"))
}

// ===== [ Public Functions ] =====
