package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"uc/internal/constant"
	"uc/pkg/jwt"
)

// JwtMiddleware 权限校验中间件
func JwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data := map[string]interface{}{}

		// 获取token Bearer+token
		authorization := ctx.GetHeader("Authorization")
		if "" == authorization {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": constant.UNAUTHORIZED,
				"msg":  constant.CodeMap[constant.UNAUTHORIZED],
				"data": data,
			})
			ctx.Abort()
			return
		}
		authorizationSplit := strings.Split(authorization, " ")
		if "Bearer" != authorizationSplit[0] {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": constant.UNAUTHORIZED,
				"msg":  constant.CodeMap[constant.UNAUTHORIZED],
				"data": data,
			})
			ctx.Abort()
			return
		}
		token := authorizationSplit[1]
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": constant.UNAUTHORIZED,
				"msg":  constant.CodeMap[constant.UNAUTHORIZED],
				"data": data,
			})
			ctx.Abort()
			return
		}
		// 解析token
		accessClaim, err := jwt.VerifyToken(token)
		if err != nil || accessClaim.UID == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": constant.FORBIDDEN,
				"msg":  constant.CodeMap[constant.FORBIDDEN],
				"data": data,
			})
			ctx.Abort()
			return
		}
		ctx.Set("uid", accessClaim.UID)
		ctx.Next()
	}
}
