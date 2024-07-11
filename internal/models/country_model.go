package models

import (
	"errors"
	"gorm.io/gorm"
	"uc/internal/mysql"
)

type Country struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ChineseName   string `json:"chinese_name"`
	StartChar     string `json:"start_char"`
	TelephoneCode string `json:"telephone_code"`
	CreateTime    int64  `json:"create_time"`
	UpdateTime    int64  `json:"update_time"`
}

// TableName 映射表名
func (c *Country) TableName() string {
	return "country"
}

// List 国家列表
func (c *Country) List() ([]Country, error) {
	var data []Country
	result := mysql.DBG.Find(&data)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return data, nil
		}
		return nil, result.Error
	}
	return data, nil
}

// FindById 根据id查询国家
func (c *Country) FindById() (*Country, error) {
	var data *Country
	err := mysql.DBG.Where("id= ?", c.ID).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return data, nil
		}
		return nil, err
	}
	return data, nil
}
