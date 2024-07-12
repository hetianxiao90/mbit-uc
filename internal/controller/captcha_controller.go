package Controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/wenlng/go-captcha-assets/helper"
	"strconv"
	"uc/internal/constant"
	"uc/pkg/captcha"
	"uc/pkg/logger"
	"uc/pkg/redis"
	"uc/pkg/util"
)

type CaptchaController struct {
	BaseController
}

type checkRequest struct {
	Point string `json:"point" binding:"required"`
	Key   string `json:"key" binding:"required"`
}

// Get 获取人机校验
func (c *CaptchaController) Get(ctx *gin.Context) {
	err, v := captcha.GetSlideBasic()
	if err != nil {
		logger.Logger.Error(err)
		c.JsonResp(ctx, constant.CAPTCHA_GET_ERROR, nil)
		return
	}
	dotsByte, _ := json.Marshal(v.BlockData)
	key := helper.StringToMD5(string(dotsByte) + strconv.FormatInt(util.RandInt64(1000, 9999), 10))
	err = redis.Client.Set(key, dotsByte)
	if err != nil {
		logger.Logger.Error(err)
		c.JsonResp(ctx, constant.CAPTCHA_GET_ERROR, nil)
		return
	}
	respData := map[string]interface{}{
		"captcha_key":  key,
		"image_base64": v.ImageBase64,
		"tile_base64":  v.TileBase64,
		"tile_width":   v.BlockData.Width,
		"tile_height":  v.BlockData.Height,
		"tile_x":       v.BlockData.TileX,
		"tile_y":       v.BlockData.TileY,
	}
	c.JsonResp(ctx, constant.SUCCESS, respData)
	return
}

// Check 校验人机校验
func (c *CaptchaController) Check(ctx *gin.Context) {
	req := new(checkRequest)
	if err := ctx.ShouldBindBodyWithJSON(req); err != nil {
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
		return
	}

	captchaData, err := redis.Client.Get(req.Key)
	if err != nil || len(captchaData) == 0 {
		c.JsonResp(ctx, constant.CAPTCHA_CHECK_ERROR, nil)
		return
	}
	err = captcha.CheckSlide(&captcha.CheckSlideData{
		Point:         req.Point,
		Key:           req.Key,
		CacheDataByte: []byte(captchaData),
	})
	if err != nil {
		c.JsonResp(ctx, constant.CAPTCHA_CHECK_ERROR, nil)
		return
	}
	err = redis.Client.Set(req.Key+constant.REDIS_CAPTCHA_PASS_KEY, true)
	if err != nil {
		logger.Logger.Error(err)
		c.JsonResp(ctx, constant.CAPTCHA_CHECK_ERROR, nil)
		return
	}
	c.JsonResp(ctx, constant.SUCCESS, nil)
	return
}
