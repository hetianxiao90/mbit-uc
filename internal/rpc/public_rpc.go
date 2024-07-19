package rpc

import (
	"context"
	"encoding/json"
	"github.com/wenlng/go-captcha-assets/helper"
	"strconv"
	"uc/internal/constant"
	"uc/internal/models"
	"uc/internal/protoc"
	"uc/pkg/captcha"
	"uc/pkg/logger"
	"uc/pkg/redis"
	"uc/pkg/util"
)

type PublicRpc struct {
	protoc.PublicServer
}

func (pr PublicRpc) GetCaptcha(ctx context.Context, req *protoc.PublicReq) (rsp *protoc.GetCaptchaRsp, err error) {
	err, v := captcha.GetSlideBasic()
	if err != nil {
		logger.Logger.Errorf("GetCaptcha captcha.GetSlideBasic err:%v", err)
		return &protoc.GetCaptchaRsp{
			Code:    constant.CAPTCHA_GET_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_GET_ERROR],
			Data:    &protoc.GetCaptchaRsp_Data{},
		}, nil
	}
	dotsByte, _ := json.Marshal(v.BlockData)
	key := helper.StringToMD5(string(dotsByte) + strconv.FormatInt(util.RandInt64(1000, 9999), 10))
	err = redis.Client.Set(key, dotsByte)
	if err != nil {
		logger.Logger.Errorf("GetCaptcha redis.Client.Set err:%v", err)
		return &protoc.GetCaptchaRsp{
			Code:    constant.CAPTCHA_GET_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_GET_ERROR],
			Data:    &protoc.GetCaptchaRsp_Data{},
		}, nil
	}
	return &protoc.GetCaptchaRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data: &protoc.GetCaptchaRsp_Data{
			CaptchaKey:  key,
			ImageBase64: v.ImageBase64,
			TileBase64:  v.TileBase64,
			TileWidth:   int32(v.BlockData.Width),
			TileHeight:  int32(v.BlockData.Height),
			TileX:       int32(v.BlockData.TileX),
			TileY:       int32(v.BlockData.TileY),
		},
	}, nil
}

func (pr PublicRpc) PostCaptcha(ctx context.Context, req *protoc.PostCaptchaReq) (rsp *protoc.PublicRsp, err error) {
	if req.Key == "" || req.Point == "" {
		return &protoc.PublicRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.PublicRsp_Data{},
		}, nil
	}

	captchaData, err := redis.Client.Get(req.Key)
	if err != nil || len(captchaData) == 0 {
		logger.Logger.Errorf("PostCaptcha redis.Client.Get err:%v", err)
		return &protoc.PublicRsp{
			Code:    constant.CAPTCHA_CHECK_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_CHECK_ERROR],
			Data:    &protoc.PublicRsp_Data{},
		}, nil
	}
	err = captcha.CheckSlide(&captcha.CheckSlideData{
		Point:         req.Point,
		Key:           req.Key,
		CacheDataByte: []byte(captchaData),
	})
	if err != nil {
		logger.Logger.Errorf("PostCaptcha captcha.CheckSlide err:%v", err)
		return &protoc.PublicRsp{
			Code:    constant.CAPTCHA_CHECK_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_CHECK_ERROR],
			Data:    &protoc.PublicRsp_Data{},
		}, nil
	}
	err = redis.Client.Set(req.Key+constant.REDIS_CAPTCHA_PASS_KEY, true)
	if err != nil {
		logger.Logger.Errorf("PostCaptcha redis.Client.Set err:%v", err)
		return &protoc.PublicRsp{
			Code:    constant.CAPTCHA_CHECK_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_CHECK_ERROR],
			Data:    &protoc.PublicRsp_Data{},
		}, nil
	}
	return &protoc.PublicRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    &protoc.PublicRsp_Data{},
	}, nil
}

func (pr PublicRpc) GetCountry(ctx context.Context, req *protoc.PublicReq) (rsp *protoc.GetCountryRsp, err error) {
	// 查询国家数据
	var country = models.Country{}
	list, err := country.List()
	if err != nil {
		logger.Logger.Errorf("GetCountry country.List err:%v", err)
		return &protoc.GetCountryRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    []*protoc.GetCountryRsp_Data{},
		}, nil
	}
	var result []*protoc.GetCountryRsp_Data
	for _, item := range list {
		result = append(result, &protoc.GetCountryRsp_Data{
			Id:            item.ID,
			Name:          item.Name,
			ChineseName:   item.ChineseName,
			StartChar:     item.StartChar,
			TelephoneCode: item.TelephoneCode,
		})
	}
	return &protoc.GetCountryRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SYSTEM_ERROR],
		Data:    result,
	}, nil
}
