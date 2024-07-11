package email

import (
	"fmt"
	"testing"
	"uc/configs"
)

func TestSendEmail(t *testing.T) {
	// 配置初始化
	configs.Init()
	// 邮箱初始化
	Init()
	err := MyEmail.SendEmail("mbit的第一封信", []string{"hetianxiao90@163.com"}, MAIL_TYPE_HTML, `<p style='color:red'>Hello World</p>`)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Email sent successfully.")
	}
}
