package Controller

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
	"uc/internal/constant"
	"uc/internal/enum"
	"uc/internal/models"
	"uc/internal/types"
	"uc/pkg/logger"
	"uc/pkg/redis"
	"uc/pkg/util"
)

type UserController struct {
	BaseController
}

// Register 注册
func (c *UserController) Register(ctx *gin.Context) {
	req := new(types.RegisterReq)
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
		return
	}

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Errorf("SendEmail util.CheckEmail err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if !isEmail {
		c.JsonResp(ctx, constant.EMAIL_FORMAT_ERROR, nil)
		return
	}
	// 校验密码格式
	isPassword, err := util.CheckPassword(req.Password)
	if err != nil {
		logger.Errorf("Register util.CheckEmail err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if !isPassword {
		c.JsonResp(ctx, constant.PASSWORD_FORMAT_ERROR, nil)
		return
	}
	// 校验国家编码
	country := models.Country{
		ID: req.CountryCode,
	}
	countryData, err := country.FindById()
	if err != nil || countryData.ID == "" {
		logger.Errorf("Register country.FindById err:%v", err)
		c.JsonResp(ctx, constant.REGISTER_COUNTRY_ERROR, nil)
		return
	}
	// 校验邮箱验证码 前端到了这步错误都是过期，暂不做程序上判断
	code, err := redis.Client.Get(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(enum.EmailRegisterCode)))
	if err != nil || code != req.VerificationCode {
		c.JsonResp(ctx, constant.REGISTER_EMAIL_CODE_EXPIRE, nil)
		return
	}
	// 邮箱校验是否已注册
	user := &models.User{
		Email: req.Email,
	}
	findUserByEmailData, err := user.FindUserByEmail()
	if err != nil {
		logger.Error("handleRegisterCode FindUserByEmail error:", err, req)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if findUserByEmailData.Email != "" {
		c.JsonResp(ctx, constant.EMAIL_EXIST, nil)
		return
	}

	salt, err := util.GenerateSalt(12)
	if err != nil {
		logger.Errorf("Register util.GenerateSalt err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password+salt), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Register bcrypt.GenerateFromPassword err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}

	// 执行注册
	var userData = &models.User{
		UID:        util.RandInt64(20000000000, 99999999999),
		Username:   util.EncryptionEmail(req.Email),
		Password:   string(hashedPassword),
		Salt:       salt,
		Email:      req.Email,
		Status:     enum.AccountStatusNormal,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}
	err = userData.Create()
	if err != nil {
		logger.Errorf("Register userData.Create err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	c.JsonResp(ctx, constant.SUCCESS, nil)
}
