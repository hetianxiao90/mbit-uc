package constant

import "uc/internal/enum"

type SendEmailOptions struct {
	Template map[string]string `yaml:"template"`
	Title    map[string]string `yaml:"title"`
}

var SendEmailOptionsData = map[enum.MessageBehavior]SendEmailOptions{
	enum.EmailRegisterCode: {
		Template: map[string]string{
			"ZH": "./template/register_code_zh.html",
			"EN": "./template/register_code_en.html",
		},
		Title: map[string]string{
			"ZH": "欢迎注册《mbit》",
			"EN": "Welcome to register《mbit》",
		},
	},
}
