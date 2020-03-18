package gin

import (
	"io/ioutil"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/gin-gonic/gin"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// DebugHandler - debug 테스트를 위한 log 출력용 Handler 구성
func DebugHandler(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("Method:", c.Request.Method)
		logger.Debug("URL:", c.Request.RequestURI)
		logger.Debug("Query:", c.Request.URL.Query())
		logger.Debug("Params:", c.Params)
		logger.Debug("Headers:", c.Request.Header)
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		logger.Debug("Body:", string(body))
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}
