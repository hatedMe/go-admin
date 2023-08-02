package middleware

import (
	"time"

	response "example.com/mod/core"
	"example.com/mod/utils"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			response.FailWithMessage("token 不能为空", c)
			c.Abort()
			return
		}

		claims, err := utils.ParseJwtToken(token, "secret")
		if err != nil {
			response.FailWithMessage("token 不合法", c)
			c.Abort()
		}

		if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
			response.FailWithMessage("token 已过期", c)
		}

		c.Next()
	}
}
