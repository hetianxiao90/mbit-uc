package Controller

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"uc/internal/constant"
	"uc/internal/enum"
	"uc/internal/models"
	"uc/internal/types"
	"uc/pkg/jwt"
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
		logger.Logger.Errorf("Register util.CheckEmail err:%v", err)
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
		logger.Logger.Errorf("Register util.CheckEmail err:%v", err)
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
		logger.Logger.Errorf("Register country.FindById err:%v", err)
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
		logger.Logger.Errorf("Register FindUserByEmail req:%v,error:%v", req, err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if findUserByEmailData.Email != "" {
		c.JsonResp(ctx, constant.EMAIL_EXIST, nil)
		return
	}

	salt, err := util.GenerateSalt(12)
	if err != nil {
		logger.Logger.Errorf("Register util.GenerateSalt err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	password := util.HashPassword(req.Password, salt)
	if err != nil {
		logger.Logger.Errorf("Register bcrypt.GenerateFromPassword err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}

	// 执行注册
	var userData = &models.User{
		UID:        util.RandInt64(20000000000, 99999999999),
		Username:   util.EncryptionEmail(req.Email),
		Password:   password,
		Salt:       salt,
		Email:      req.Email,
		Status:     enum.AccountStatusNormal,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}
	err = userData.Create()
	if err != nil {
		logger.Logger.Errorf("Register userData.Create err: %v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	// 注册通过获取Token
	token, refreshToken, err := jwt.CreateToken(userData.UID)
	if err != nil {
		logger.Logger.Errorf("Register jwt.CreateToken err: %v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	loginResult := types.LoginResult{
		UID:          userData.UID,
		Username:     userData.Username,
		Email:        util.EncryptionEmail(userData.Email),
		AccessToken:  token,
		RefreshToken: refreshToken,
	}
	c.JsonResp(ctx, constant.SUCCESS, loginResult)
}

func (c *UserController) Login(ctx *gin.Context) {
	req := new(types.LoginReq)
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
		return
	}

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Login util.CheckEmail err:%v", err)
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
		logger.Logger.Errorf("Login util.CheckEmail err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if !isPassword {
		c.JsonResp(ctx, constant.PASSWORD_FORMAT_ERROR, nil)
		return
	}
	// 查找用户
	user := &models.User{
		Email: req.Email,
	}
	findUserByEmailData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Errorf("Login FindUserByEmail req:%v,error:%v", req, err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	// 用户不存在
	if findUserByEmailData.Email == "" {
		c.JsonResp(ctx, constant.LOGIN_USER_IS_NOT_EXIST, nil)
		return
	}
	reqPassword := util.HashPassword(req.Password, findUserByEmailData.Salt)

	if findUserByEmailData.Password != reqPassword {
		c.JsonResp(ctx, constant.LOGIN_USER_PASSWORD_ERROR, nil)
		return
	}
	// 登录通过获取Token TODO 要做退出登录需要存入redis
	token, refreshToken, err := jwt.CreateToken(findUserByEmailData.UID)
	if err != nil {
		logger.Logger.Errorf("Login jwt.CreateToken err: %v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	loginResult := types.LoginResult{
		UID:          findUserByEmailData.UID,
		Username:     findUserByEmailData.Username,
		Email:        util.EncryptionEmail(findUserByEmailData.Email),
		AccessToken:  token,
		RefreshToken: refreshToken,
	}
	c.JsonResp(ctx, constant.SUCCESS, loginResult)
}

func (c *UserController) RefreshToken(ctx *gin.Context) {

	// 获取refreshToken
	refreshToken := ctx.GetHeader("refresh-token")
	newAccessToken, newRefreshToken, err := jwt.RefreshToken(refreshToken)
	if err != nil {
		c.JsonResp(ctx, constant.REFRESH_TOKEN_FAILED, nil)
		return
	}
	result := types.RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	c.JsonResp(ctx, constant.SUCCESS, result)
}

// Info 用户信息
func (c *UserController) Info(ctx *gin.Context) {
	uid, _ := ctx.Get("uid")
	user := models.User{
		UID: uid.(int64),
	}
	result, err := user.FindUserByUid()
	if err != nil {
		logger.Logger.Errorf("Info user.FindUserByUid err:%v", err)
		c.JsonResp(ctx, constant.USER_NOT_EXIST, nil)
		return
	}
	c.JsonResp(ctx, constant.SUCCESS, result)
}
