package types

type RegisterReq struct {
	CountryCode      string `json:"country_code" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Password         string `json:"password" binding:"required"`
	VerificationCode string `json:"verification_code" binding:"required"` // 验证码
}

type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	UID          int64  `json:"uid"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	AccessToken  string `json:"assess_token"`
	RefreshToken string `json:"refresh_token"`
}
type RefreshTokenResult struct {
	AccessToken  string `json:"assess_token"`
	RefreshToken string `json:"refresh_token"`
}
