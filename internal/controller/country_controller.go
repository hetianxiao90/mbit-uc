package Controller

import (
	"github.com/gin-gonic/gin"
	"uc/internal/constant"
	"uc/internal/models"
	"uc/pkg/logger"
)

type CountryController struct {
	BaseController
}

// List 获取人机校验
func (c *CountryController) List(ctx *gin.Context) {
	// 查询国家数据
	var country = models.Country{}
	list, err := country.List()
	if err != nil {
		logger.Logger.Errorf("CountryController List err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, list)
		return
	}
	c.JsonResp(ctx, constant.SUCCESS, list)
	return
}
