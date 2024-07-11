package types

type RegisterReq struct {
	CountryCode      string `json:"country_code" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Password         string `json:"password" binding:"required"`
	VerificationCode string `json:"verification_code" binding:"required"` // 验证码
}
