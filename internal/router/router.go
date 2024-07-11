package router

import (
	"github.com/gin-gonic/gin"
	routerV1 "uc/internal/router/v1"
)

func Init() *gin.Engine {
	r := gin.Default()

	routerV1.UserRouter(r)
	routerV1.PublicRouter(r)
	return r
}
