package enum

// AccountStatus 账号状态
type AccountStatus int8

const (
	AccountStatusNormal    AccountStatus = 1 // 正常
	AccountStatusForbidden AccountStatus = 2 // 冻结
)

// MessageBehavior 消息行为
type MessageBehavior int16

const (
	EmailRegisterCode MessageBehavior = 1001 // 发送注册邮箱
)
