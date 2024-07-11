package models

import (
	"errors"
	"gorm.io/gorm"
	"uc/internal/enum"
	"uc/internal/mysql"
)

type User struct {
	UID        int64              `json:"uid"`
	Username   string             `json:"username"`
	Password   string             `json:"password"`
	Salt       string             `json:"salt"`
	Email      string             `json:"email"`
	Status     enum.AccountStatus `json:"status"`
	CreateTime int64              `json:"create_time"`
	UpdateTime int64              `json:"update_time"`
}

// TableName 映射表名
func (u *User) TableName() string {
	return "user"
}

// FindUserByEmail 根据email查找user
func (u *User) FindUserByEmail() (*User, error) {
	user := &User{}
	err := mysql.DBG.Where("email = ?", u.Email).Select("email").First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {

		return nil, err
	}
	return user, nil
}

// Create 创建
func (u *User) Create() error {
	err := mysql.DBG.Create(u).Error
	return err
}
