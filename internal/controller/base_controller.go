package Controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"uc/internal/constant"
)

type BaseController struct{}

func (c BaseController) JsonResp(ctx *gin.Context, errCode int, data interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	resp := gin.H{
		"code": errCode,
		"msg":  constant.CodeMap[errCode],
		"data": data,
	}
	ctx.JSON(http.StatusOK, resp)
	return
}
