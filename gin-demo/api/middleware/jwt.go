package middleware

import (
	"errors"
	"gin-demo/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var Secret = []byte("114514")

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.RequestURI, "login") {
			return
		} else if strings.Contains(c.Request.RequestURI, "register") {
			return
		}

		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"msg": "请求头中auth为空",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"msg": "请求头中auth格式有误",
			})
			c.Abort()
			return
		}

		mc, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"msg": "授权已过期,请重新登录",
			})
			c.Abort()
			return
		}

		c.Set("account", mc.Account)
		c.Next()
	}
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*model.MyClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &model.MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return Secret, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.MyClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
