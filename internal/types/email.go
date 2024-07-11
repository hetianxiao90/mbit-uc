package types

import "uc/internal/enum"

type SendEmailCodeReq struct {
	Email    string               `form:"email" json:"email" binding:"required"`
	Key      string               `form:"key" json:"key"` // 人机校验key
	Behavior enum.MessageBehavior `form:"behavior" json:"behavior" binding:"required"`
}

type SendEmailData struct {
	Email    string               `json:"email"`
	Behavior enum.MessageBehavior `json:"behavior"`
	Language string               `json:"language"`
	Data     string               `json:"data"`
}

type CheckEmailReq struct {
	Email    string               `form:"email" json:"email" binding:"required"`
	Behavior enum.MessageBehavior `form:"behavior" json:"behavior" binding:"required"`
	Code     string               `form:"code" json:"code" binding:"required"`
}
