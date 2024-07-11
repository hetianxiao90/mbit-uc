package constant

// uc 10000-11000
const (
	SUCCESS                    = 200   // 成功
	SYSTEM_ERROR               = 500   // 成功
	ENTITY_ERROR               = 422   // 参数错误
	CAPTCHA_GET_ERROR          = 10000 // 获取人机校验数据失败
	CAPTCHA_CHECK_ERROR        = 10001 // 人机校验校验失败
	EMAIL_EXIST                = 10002 // 人机校验校验失败
	EMAIL_FORMAT_ERROR         = 10003 // 邮箱错误
	REGISTER_EMAIL_CODE_ERROR  = 10004 // 注册邮箱验证码错误
	PASSWORD_FORMAT_ERROR      = 10005 // 邮箱错误
	REGISTER_COUNTRY_ERROR     = 10006 // 国家编码错误
	REGISTER_EMAIL_CODE_EXPIRE = 10007 // 注册邮箱验证码过期
)

var CodeMap = map[int]string{
	SUCCESS:                    "success",
	SYSTEM_ERROR:               "system error",
	ENTITY_ERROR:               "entity error",
	CAPTCHA_GET_ERROR:          "captcha get error",
	CAPTCHA_CHECK_ERROR:        "captcha check error",
	EMAIL_EXIST:                "email exist",
	EMAIL_FORMAT_ERROR:         "email format error",
	REGISTER_EMAIL_CODE_ERROR:  "register email code error",
	PASSWORD_FORMAT_ERROR:      "password format error",
	REGISTER_COUNTRY_ERROR:     "register country error",
	REGISTER_EMAIL_CODE_EXPIRE: "register email code expire",
}
