package rpc

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
	"uc/internal/constant"
	"uc/internal/enum"
	"uc/internal/models"
	"uc/internal/protoc"
	"uc/internal/types"
	"uc/pkg/jwt"
	"uc/pkg/logger"
	"uc/pkg/nacos"
	"uc/pkg/rabbitmq"
	"uc/pkg/redis"
	"uc/pkg/util"
)

type UserRpc struct {
	protoc.UcServer
}

func (ur UserRpc) GetEmailCode(ctx context.Context, req *protoc.GetEmailCodeReq) (rsp *protoc.UcRsp, err error) {

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("GetRegisterEmailCode util.CheckEmail err:%v", err)
		return &protoc.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	if !isEmail {
		return &protoc.UcRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	// 区分行为
	switch req.Behavior {
	case int32(enum.EmailRegisterCode):
		return handleSendRegisterCode(ctx, req)
	default:
		return &protoc.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
}
func handleSendRegisterCode(ctx context.Context, req *protoc.GetEmailCodeReq) (rsp *protoc.UcRsp, err error) {
	// 参数校验
	key := req.Key
	if key == "" {
		return &protoc.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	// 人机校验是否成功
	captchaResult, err := redis.Client.Get(req.Key + constant.REDIS_CAPTCHA_PASS_KEY)
	if err != nil || captchaResult != "true" {
		return &protoc.UcRsp{
			Code:    constant.CAPTCHA_CHECK_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_CHECK_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}

	// 邮箱校验是否已注册
	user := &models.User{
		Email: req.Email,
	}
	userData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Error("handleSendRegisterCode FindUserByEmail error:", err, req)
		return &protoc.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	if userData.Email != "" {
		return &protoc.UcRsp{
			Code:    constant.EMAIL_EXIST,
			Message: constant.CodeMap[constant.EMAIL_EXIST],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	//生成code
	code := util.RandInt64(100000, 999999)
	err = redis.Client.Set(req.Email+constant.REDIS_EMAIL_SEND_REGISTER_CODE+strconv.Itoa(int(req.Behavior)), code)
	if err != nil {
		logger.Logger.Errorf("handleRegisterCode redis.Client.Set: %v", err)
		return &protoc.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	// mq消息
	amqpData := types.SendEmailData{
		Behavior: enum.MessageBehavior(req.Behavior),
		Language: "ZH",
		Email:    req.Email,
		Data:     strconv.FormatInt(code, 10),
	}
	//传入mq
	amqpDataJson, _ := json.Marshal(amqpData)
	err = rabbitmq.AMQP.Publish(
		nacos.Config.RabbitMq.Exchanges.User,
		nacos.Config.RabbitMq.RoutingKey.Public,
		amqpDataJson,
	)
	if err != nil {
		logger.Logger.Errorf("handleRegisterCode rabbitmq.AMQP.Publish: %v", err)
		return &protoc.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	return &protoc.UcRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    &protoc.UcRsp_Data{},
	}, nil
}

func (ur UserRpc) PostEmailCode(ctx context.Context, req *protoc.PostEmailCodeReq) (rsp *protoc.UcRsp, err error) {
	// 参数校验
	if req.Email == "" || req.Code == "" || req.Behavior == 0 {
		return &protoc.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("PostEmailCode util.CheckEmail err:%v", err)
		return &protoc.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	if !isEmail {
		return &protoc.UcRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	// 区分行为
	switch req.Behavior {
	case int32(enum.EmailRegisterCode):
		return handleCheckRegisterCode(ctx, req)

	default:
		return &protoc.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
}

func handleCheckRegisterCode(ctx context.Context, req *protoc.PostEmailCodeReq) (rsp *protoc.UcRsp, err error) {
	code, err := redis.Client.Get(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(req.Behavior)))
	if err != nil || code != req.Code {
		return &protoc.UcRsp{
			Code:    constant.REGISTER_EMAIL_CODE_ERROR,
			Message: constant.CodeMap[constant.REGISTER_EMAIL_CODE_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	// 更新code过期时间
	err = redis.Client.Expire(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(req.Behavior)))
	if err != nil {
		return &protoc.UcRsp{
			Code:    constant.REGISTER_EMAIL_CODE_ERROR,
			Message: constant.CodeMap[constant.REGISTER_EMAIL_CODE_ERROR],
			Data:    &protoc.UcRsp_Data{},
		}, nil
	}
	return &protoc.UcRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    &protoc.UcRsp_Data{},
	}, nil
}

func (ur UserRpc) Register(ctx context.Context, req *protoc.RegisterReq) (rsp *protoc.LoginRsp, err error) {
	// 参数校验
	if req.Email == "" || req.CountryId == "" || req.Password == "" || req.VerificationCode == "" {
		return &protoc.LoginRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Register util.CheckEmail err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	if !isEmail {
		return &protoc.LoginRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 校验密码格式
	isPassword, err := util.CheckPassword(req.Password)
	if err != nil {
		logger.Logger.Errorf("Register util.CheckEmail err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	if !isPassword {
		return &protoc.LoginRsp{
			Code:    constant.PASSWORD_FORMAT_ERROR,
			Message: constant.CodeMap[constant.PASSWORD_FORMAT_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 校验国家编码
	country := models.Country{
		ID: req.CountryId,
	}
	countryData, err := country.FindById()
	if err != nil || countryData.ID == "" {
		logger.Logger.Errorf("Register country.FindById err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.REGISTER_COUNTRY_ERROR,
			Message: constant.CodeMap[constant.REGISTER_COUNTRY_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 校验邮箱验证码 前端到了这报错都是过期，暂不做程序上判断
	code, err := redis.Client.Get(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(enum.EmailRegisterCode)))
	if err != nil || code != req.VerificationCode {
		return &protoc.LoginRsp{
			Code:    constant.REGISTER_EMAIL_CODE_EXPIRE,
			Message: constant.CodeMap[constant.REGISTER_EMAIL_CODE_EXPIRE],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 邮箱校验是否已注册
	user := &models.User{
		Email: req.Email,
	}
	findUserByEmailData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Errorf("Register FindUserByEmail req:%v,error:%v", req, err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil

	}
	if findUserByEmailData.Email != "" {
		return &protoc.LoginRsp{
			Code:    constant.EMAIL_EXIST,
			Message: constant.CodeMap[constant.EMAIL_EXIST],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}

	salt, err := util.GenerateSalt(12)
	if err != nil {
		logger.Logger.Errorf("Register util.GenerateSalt err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	password := util.HashPassword(req.Password, salt)
	if err != nil {
		logger.Logger.Errorf("Register bcrypt.GenerateFromPassword err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}

	// 执行注册
	var userData = &models.User{
		UID:        util.RandInt64(20000000000, 99999999999),
		Username:   util.EncryptionEmail(req.Email),
		Password:   password,
		Salt:       salt,
		Email:      req.Email,
		CountryId:  req.CountryId,
		Status:     enum.AccountStatusNormal,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}
	err = userData.Create()
	if err != nil {
		logger.Logger.Errorf("Register userData.Create err: %v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 注册通过获取Token
	token, refreshToken, err := jwt.CreateToken(userData.UID)
	if err != nil {
		logger.Logger.Errorf("Register jwt.CreateToken err: %v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	result := &protoc.LoginRsp_Data{
		Uid:          userData.UID,
		Username:     userData.Username,
		Email:        util.EncryptionEmail(userData.Email),
		AccessToken:  token,
		RefreshToken: refreshToken,
	}
	return &protoc.LoginRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    result,
	}, nil
}

func (ur UserRpc) Login(ctx context.Context, req *protoc.LoginReq) (rsp *protoc.LoginRsp, err error) {
	// 参数校验
	if req.Email == "" || req.Password == "" {
		return &protoc.LoginRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Login util.CheckEmail err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	if !isEmail {
		return &protoc.LoginRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 校验密码格式
	isPassword, err := util.CheckPassword(req.Password)
	if err != nil {
		logger.Logger.Errorf("Login util.CheckEmail err:%v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	if !isPassword {
		return &protoc.LoginRsp{
			Code:    constant.PASSWORD_FORMAT_ERROR,
			Message: constant.CodeMap[constant.PASSWORD_FORMAT_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 查找用户
	user := &models.User{
		Email: req.Email,
	}
	findUserByEmailData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Errorf("Login FindUserByEmail req:%v,error:%v", req, err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 用户不存在
	if findUserByEmailData.Email == "" {
		return &protoc.LoginRsp{
			Code:    constant.LOGIN_USER_IS_NOT_EXIST,
			Message: constant.CodeMap[constant.LOGIN_USER_IS_NOT_EXIST],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	reqPassword := util.HashPassword(req.Password, findUserByEmailData.Salt)

	if findUserByEmailData.Password != reqPassword {
		return &protoc.LoginRsp{
			Code:    constant.LOGIN_USER_PASSWORD_ERROR,
			Message: constant.CodeMap[constant.LOGIN_USER_PASSWORD_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	// 登录通过获取Token TODO 要做退出登录需要存入redis
	accessToken, refreshToken, err := jwt.CreateToken(findUserByEmailData.UID)
	if err != nil {
		logger.Logger.Errorf("Login jwt.CreateToken err: %v", err)
		return &protoc.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
			Data:    &protoc.LoginRsp_Data{},
		}, nil
	}
	result := &protoc.LoginRsp_Data{
		Uid:          findUserByEmailData.UID,
		Username:     findUserByEmailData.Username,
		Email:        util.EncryptionEmail(findUserByEmailData.Email),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return &protoc.LoginRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    result,
	}, nil
}

func (ur UserRpc) GetUserInfo(ctx context.Context, req *protoc.GetUserInfoReq) (rsp *protoc.GetUserInfoRsp, err error) {

	// 参数校验
	if req.Uid == 0 {
		return &protoc.GetUserInfoRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
			Data:    &protoc.GetUserInfoRsp_Data{},
		}, nil
	}

	user := models.User{
		UID: req.Uid,
	}
	userData, err := user.FindUserByUid()
	if err != nil {
		logger.Logger.Errorf("Info user.FindUserByUid err:%v", err)
		return &protoc.GetUserInfoRsp{
			Code:    constant.USER_NOT_EXIST,
			Message: constant.CodeMap[constant.USER_NOT_EXIST],
		}, nil
	}
	result := &protoc.GetUserInfoRsp_Data{
		Uid:       userData.UID,
		Username:  userData.Username,
		Email:     util.EncryptionEmail(userData.Email),
		CountryId: userData.CountryId,
	}
	return &protoc.GetUserInfoRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    result,
	}, nil
}
