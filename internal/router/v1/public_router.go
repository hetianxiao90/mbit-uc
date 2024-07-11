package v1

import (
	"github.com/gin-gonic/gin"
	"uc/internal/controller"
)

func PublicRouter(e *gin.Engine) {
	r := e.Group("v1/public")
	{
		// 验证码
		captcha := Controller.CaptchaController{}
		r.GET("/captcha/", captcha.Get)
		r.POST("/captcha/", captcha.Check)

		// 国家
		country := Controller.CountryController{}
		r.GET("/country/", country.List)
	}
}
