package Controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
	"uc/configs"
	"uc/internal/constant"
	"uc/internal/enum"
	"uc/internal/logger"
	"uc/internal/models"
	"uc/internal/rabbitmq"
	"uc/internal/redis"
	"uc/internal/types"
	"uc/internal/util"
)

type EmailController struct {
	BaseController
}

func (c *EmailController) SendCode(ctx *gin.Context) {
	// 参数接收
	req := new(types.SendEmailCodeReq)
	if err := ctx.ShouldBindQuery(&req); err != nil {
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
	// 区分行为
	switch req.Behavior {
	case enum.EmailRegisterCode:
		c.handleSendRegisterCode(ctx, req)
		return
	default:
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
	}
}

// 注册验证码
func (c *EmailController) handleSendRegisterCode(ctx *gin.Context, req *types.SendEmailCodeReq) {
	// 参数校验
	key := req.Key
	if key == "" {
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
		return
	}
	// 人机校验是否成功
	captchaResult, err := redis.Client.Get(req.Key + constant.REDIS_CAPTCHA_PASS_KEY)
	if err != nil || captchaResult != "true" {
		c.JsonResp(ctx, constant.CAPTCHA_CHECK_ERROR, nil)
		return
	}

	// 邮箱校验是否已注册
	user := &models.User{
		Email: req.Email,
	}
	userData, err := user.FindUserByEmail()
	if err != nil {
		logger.Error("handleRegisterCode FindUserByEmail error:", err, req)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if userData.Email != "" {
		c.JsonResp(ctx, constant.EMAIL_EXIST, nil)
		return
	}
	//生成code
	code := util.RandInt64(100000, 999999)
	err = redis.Client.Set(req.Email+constant.REDIS_EMAIL_SEND_REGISTER_CODE+strconv.Itoa(int(req.Behavior)), code)
	if err != nil {
		logger.Errorf("handleRegisterCode redis.Client.Set: %v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	// mq消息
	amqpData := types.SendEmailData{
		Behavior: req.Behavior,
		Language: "ZH",
		Email:    req.Email,
		Data:     strconv.FormatInt(code, 10),
	}
	//传入mq
	amqpDataJson, _ := json.Marshal(amqpData)
	err = rabbitmq.AMQP.Publish(
		configs.Config.RabbitMq.Exchanges.User,
		configs.Config.RabbitMq.RoutingKey.Public,
		amqpDataJson,
	)
	if err != nil {
		logger.Errorf("handleRegisterCode rabbitmq.AMQP.Publish: %v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	c.JsonResp(ctx, constant.SUCCESS, nil)
}

// CheckCode 校验邮箱code
func (c *EmailController) CheckCode(ctx *gin.Context) {
	// 参数接收
	req := new(types.CheckEmailReq)
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
		return
	}
	// 校验邮箱格式
	isEmail, err := util.CheckEmail(req.Email)
	if err != nil {
		logger.Errorf("CheckEmailCode util.CheckEmail err:%v", err)
		c.JsonResp(ctx, constant.SYSTEM_ERROR, nil)
		return
	}
	if !isEmail {
		c.JsonResp(ctx, constant.EMAIL_FORMAT_ERROR, nil)
		return
	}
	// 区分行为
	switch req.Behavior {
	case enum.EmailRegisterCode:
		c.handleCheckRegisterCode(ctx, req)
		return
	default:
		c.JsonResp(ctx, constant.ENTITY_ERROR, nil)
	}

}

func (c *EmailController) handleCheckRegisterCode(ctx *gin.Context, req *types.CheckEmailReq) {
	code, err := redis.Client.Get(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(req.Behavior)))
	if err != nil || code != req.Code {
		c.JsonResp(ctx, constant.REGISTER_EMAIL_CODE_ERROR, nil)
		return
	}
	// 更新code过期时间
	err = redis.Client.Expire(req.Email + constant.REDIS_EMAIL_SEND_REGISTER_CODE + strconv.Itoa(int(req.Behavior)))
	if err != nil {
		c.JsonResp(ctx, constant.REGISTER_EMAIL_CODE_ERROR, nil)
		return
	}
	c.JsonResp(ctx, constant.SUCCESS, nil)
}
