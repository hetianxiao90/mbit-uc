package rpc

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
	"uc/internal/constant"
	"uc/internal/enum"
	"uc/internal/models"
	proto "uc/internal/protoc"
	"uc/internal/types"
	"uc/pkg/jwt"
	"uc/pkg/logger"
	"uc/pkg/nacos"
	"uc/pkg/rabbitmq"
	"uc/pkg/redis"
	"uc/pkg/util"
)

type UserRpc struct {
	proto.UcServer
}

func (ur UserRpc) GetEmailCode(ctx context.Context, req *proto.GetEmailCodeReq) (rsp *proto.UcRsp, err error) {

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("GetRegisterEmailCode util.CheckEmail err:%v", err)
		return &proto.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if !isEmail {
		return &proto.UcRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
		}, nil
	}
	// 区分行为
	switch req.Behavior {
	case int32(enum.EmailRegisterCode):
		return handleSendRegisterCode(ctx, req)
	default:
		return &proto.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
		}, nil
	}
}
func handleSendRegisterCode(ctx context.Context, req *proto.GetEmailCodeReq) (rsp *proto.UcRsp, err error) {
	// 参数校验
	key := req.Key
	if key == "" {
		return &proto.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
		}, nil
	}
	// 人机校验是否成功
	captchaResult, err := redis.Client.Get(req.Key + constant.REDIS_CAPTCHA_PASS_KEY)
	if err != nil || captchaResult != "true" {
		return &proto.UcRsp{
			Code:    constant.CAPTCHA_CHECK_ERROR,
			Message: constant.CodeMap[constant.CAPTCHA_CHECK_ERROR],
		}, nil
	}

	// 邮箱校验是否已注册
	user := &models.User{
		Email: req.Email,
	}
	userData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Error("handleSendRegisterCode FindUserByEmail error:", err, req)
		return &proto.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if userData.Email != "" {
		return &proto.UcRsp{
			Code:    constant.EMAIL_EXIST,
			Message: constant.CodeMap[constant.EMAIL_EXIST],
		}, nil
	}
	//生成code
	code := util.RandInt64(100000, 999999)
	err = redis.Client.Set(req.Email+constant.REDIS_EMAIL_SEND_REGISTER_CODE+strconv.Itoa(int(req.Behavior)), code)
	if err != nil {
		logger.Logger.Errorf("handleRegisterCode redis.Client.Set: %v", err)
		return &proto.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
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
		return &proto.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	return &proto.UcRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
	}, nil
}

func (ur UserRpc) PostEmailCode(ctx context.Context, req *proto.PostEmailCodeReq) (rsp *proto.UcRsp, err error) {

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("PostEmailCode util.CheckEmail err:%v", err)
		return &proto.UcRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if !isEmail {
		return &proto.UcRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
		}, nil
	}
	// 区分行为
	switch req.Behavior {
	case int32(enum.EmailRegisterCode):
		return handleCheckRegisterCode(ctx, req)

	default:
		return &proto.UcRsp{
			Code:    constant.ENTITY_ERROR,
			Message: constant.CodeMap[constant.ENTITY_ERROR],
		}, nil
	}
}

func handleCheckRegisterCode(ctx context.Context, req *proto.PostEmailCodeReq) (rsp *proto.UcRsp, err error) {
	code, err := redis.Client.Get(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(req.Behavior)))
	if err != nil || code != req.Code {
		return &proto.UcRsp{
			Code:    constant.REGISTER_EMAIL_CODE_ERROR,
			Message: constant.CodeMap[constant.REGISTER_EMAIL_CODE_ERROR],
		}, nil
	}
	// 更新code过期时间
	err = redis.Client.Expire(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(req.Behavior)))
	if err != nil {
		return &proto.UcRsp{
			Code:    constant.REGISTER_EMAIL_CODE_ERROR,
			Message: constant.CodeMap[constant.REGISTER_EMAIL_CODE_ERROR],
		}, nil
	}
	return &proto.UcRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
	}, nil
}

func (ur UserRpc) Register(ctx context.Context, req *proto.RegisterReq) (rsp *proto.LoginRsp, err error) {

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Register util.CheckEmail err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if !isEmail {
		return &proto.LoginRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
		}, nil
	}
	// 校验密码格式
	isPassword, err := util.CheckPassword(req.Password)
	if err != nil {
		logger.Logger.Errorf("Register util.CheckEmail err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if !isPassword {
		return &proto.LoginRsp{
			Code:    constant.PASSWORD_FORMAT_ERROR,
			Message: constant.CodeMap[constant.PASSWORD_FORMAT_ERROR],
		}, nil
	}
	// 校验国家编码
	country := models.Country{
		ID: req.CountryId,
	}
	countryData, err := country.FindById()
	if err != nil || countryData.ID == "" {
		logger.Logger.Errorf("Register country.FindById err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.REGISTER_COUNTRY_ERROR,
			Message: constant.CodeMap[constant.REGISTER_COUNTRY_ERROR],
		}, nil
	}
	// 校验邮箱验证码 前端到了这报错都是过期，暂不做程序上判断
	code, err := redis.Client.Get(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(enum.EmailRegisterCode)))
	if err != nil || code != req.VerificationCode {
		return &proto.LoginRsp{
			Code:    constant.REGISTER_EMAIL_CODE_EXPIRE,
			Message: constant.CodeMap[constant.REGISTER_EMAIL_CODE_EXPIRE],
		}, nil
	}
	// 邮箱校验是否已注册
	user := &models.User{
		Email: req.Email,
	}
	findUserByEmailData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Errorf("Register FindUserByEmail req:%v,error:%v", req, err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil

	}
	if findUserByEmailData.Email != "" {
		return &proto.LoginRsp{
			Code:    constant.EMAIL_EXIST,
			Message: constant.CodeMap[constant.EMAIL_EXIST],
		}, nil
	}

	salt, err := util.GenerateSalt(12)
	if err != nil {
		logger.Logger.Errorf("Register util.GenerateSalt err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	password := util.HashPassword(req.Password, salt)
	if err != nil {
		logger.Logger.Errorf("Register bcrypt.GenerateFromPassword err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
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
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	// 注册通过获取Token
	token, refreshToken, err := jwt.CreateToken(userData.UID)
	if err != nil {
		logger.Logger.Errorf("Register jwt.CreateToken err: %v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil

	}
	result := &proto.LoginRsp_Data{
		Uid:          userData.UID,
		Username:     userData.Username,
		Email:        util.EncryptionEmail(userData.Email),
		AccessToken:  token,
		RefreshToken: refreshToken,
	}
	return &proto.LoginRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    result,
	}, nil
}

func (ur UserRpc) Login(ctx context.Context, req *proto.LoginReq) (rsp *proto.LoginRsp, err error) {

	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Login util.CheckEmail err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if !isEmail {
		return &proto.LoginRsp{
			Code:    constant.EMAIL_FORMAT_ERROR,
			Message: constant.CodeMap[constant.EMAIL_FORMAT_ERROR],
		}, nil
	}
	// 校验密码格式
	isPassword, err := util.CheckPassword(req.Password)
	if err != nil {
		logger.Logger.Errorf("Login util.CheckEmail err:%v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	if !isPassword {
		return &proto.LoginRsp{
			Code:    constant.PASSWORD_FORMAT_ERROR,
			Message: constant.CodeMap[constant.PASSWORD_FORMAT_ERROR],
		}, nil
	}
	// 查找用户
	user := &models.User{
		Email: req.Email,
	}
	findUserByEmailData, err := user.FindUserByEmail()
	if err != nil {
		logger.Logger.Errorf("Login FindUserByEmail req:%v,error:%v", req, err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	// 用户不存在
	if findUserByEmailData.Email == "" {
		return &proto.LoginRsp{
			Code:    constant.LOGIN_USER_IS_NOT_EXIST,
			Message: constant.CodeMap[constant.LOGIN_USER_IS_NOT_EXIST],
		}, nil
	}
	reqPassword := util.HashPassword(req.Password, findUserByEmailData.Salt)

	if findUserByEmailData.Password != reqPassword {
		return &proto.LoginRsp{
			Code:    constant.LOGIN_USER_PASSWORD_ERROR,
			Message: constant.CodeMap[constant.LOGIN_USER_PASSWORD_ERROR],
		}, nil
	}
	// 登录通过获取Token TODO 要做退出登录需要存入redis
	accessToken, refreshToken, err := jwt.CreateToken(findUserByEmailData.UID)
	if err != nil {
		logger.Logger.Errorf("Login jwt.CreateToken err: %v", err)
		return &proto.LoginRsp{
			Code:    constant.SYSTEM_ERROR,
			Message: constant.CodeMap[constant.SYSTEM_ERROR],
		}, nil
	}
	result := &proto.LoginRsp_Data{
		Uid:          findUserByEmailData.UID,
		Username:     findUserByEmailData.Username,
		Email:        util.EncryptionEmail(findUserByEmailData.Email),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return &proto.LoginRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    result,
	}, nil
}

func (ur UserRpc) GetUserInfo(ctx context.Context, req *proto.UcReq) (rsp *proto.GetUserInfoRsp, err error) {
	uid := ctx.Value("uid")
	user := models.User{
		UID: uid.(int64),
	}
	userData, err := user.FindUserByUid()
	if err != nil {
		logger.Logger.Errorf("Info user.FindUserByUid err:%v", err)
		return &proto.GetUserInfoRsp{
			Code:    constant.USER_NOT_EXIST,
			Message: constant.CodeMap[constant.USER_NOT_EXIST],
		}, nil
	}
	result := &proto.GetUserInfoRsp_Data{
		Uid:       userData.UID,
		Username:  userData.Username,
		Email:     util.EncryptionEmail(userData.Email),
		CountryId: userData.CountryId,
	}
	return &proto.GetUserInfoRsp{
		Code:    constant.SUCCESS,
		Message: constant.CodeMap[constant.SUCCESS],
		Data:    result,
	}, nil
}
