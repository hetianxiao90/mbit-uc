package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
	"uc/configs"
)

type AccessClaim struct {
	UID int64 `json:"uid"`
	jwt.RegisteredClaims
}
type RefreshClaim struct {
	UID int64 `json:"uid"`
	jwt.RegisteredClaims
}
type MyJwt struct {
	AccessTokenExpiredTime  int64  `json:"access_token_expired_time"`
	RefreshTokenExpiredTime int64  `json:"refresh_token_expired_time"`
	Secret                  []byte `json:"secret"`
}

var JWT *MyJwt

func Init() {
	JWT = NewMyJwt()
}

func NewMyJwt() *MyJwt {
	return &MyJwt{
		AccessTokenExpiredTime:  configs.Config.Jwt.AccessTokenExpiredTime,
		RefreshTokenExpiredTime: configs.Config.Jwt.RefreshTokenExpiredTime,
		Secret:                  []byte(configs.Config.Jwt.Secret),
	}
}
func getExpiredTime(hour int64) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(time.Duration(hour) * time.Minute))
}
func CreateToken(uid int64) (accessToken, refreshToken string, err error) {
	accessRC := jwt.RegisteredClaims{
		ExpiresAt: getExpiredTime(JWT.AccessTokenExpiredTime),
	}
	refreshRc := jwt.RegisteredClaims{
		ExpiresAt: getExpiredTime(JWT.RefreshTokenExpiredTime),
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &AccessClaim{uid, accessRC}).SignedString(JWT.Secret)
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &RefreshClaim{uid, refreshRc}).SignedString(JWT.Secret)
	return
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	return JWT.Secret, nil
}

// VerifyToken 解析token
func VerifyToken(accessToken string) (*AccessClaim, error) {
	var accessClaim = new(AccessClaim)
	jwtToken, err := jwt.ParseWithClaims(accessToken, accessClaim, keyFunc)
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		err = errors.New("jwtToken Valid Failed")
		return nil, err
	}
	return accessClaim, nil
}

func RefreshToken(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	// refreshToken 出错直接返回
	var refreshClaim = new(RefreshClaim)
	jwtToken, err := jwt.ParseWithClaims(refreshToken, refreshClaim, keyFunc)
	if err != nil {
		return
	}
	if !jwtToken.Valid {
		err = errors.New("jwtToken Valid Failed")
		return
	}
	return CreateToken(refreshClaim.UID)
}
