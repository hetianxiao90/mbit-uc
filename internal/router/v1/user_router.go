package v1

import (
	"github.com/gin-gonic/gin"
	"uc/internal/controller"
	"uc/internal/router/middleware"
)

func UserRouter(e *gin.Engine) {
	r := e.Group("v1/uc/")
	{
		user := Controller.UserController{}
		email := Controller.EmailController{}
		r.GET("/email/code", email.SendCode)
		r.POST("/email/code", email.CheckCode)
		r.POST("/register", user.Register)
		r.POST("/login", user.Login)
		r.GET("/refresh_token", user.RefreshToken)

		// 需要鉴权的接口
		ur := r.Use(middleware.JwtMiddleware())
		{
			ur.GET("/user", user.Info)

		}
	}
}
