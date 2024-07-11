package v1

import (
	"github.com/gin-gonic/gin"
	"uc/internal/controller"
)

func UserRouter(e *gin.Engine) {
	r := e.Group("v1/uc/")
	{
		user := Controller.UserController{}
		email := Controller.EmailController{}
		r.GET("/email/code", email.SendCode)
		r.POST("/email/code", email.CheckCode)
		r.POST("/register", user.Register)
		ur := r.Group("/user")
		{
			ur.GET("/email/code", email.SendCode)
		}
	}
}
