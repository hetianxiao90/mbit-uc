package rabbitmq

import (
	"bytes"
	"encoding/json"
	"html/template"
	"uc/internal/constant"
	"uc/internal/types"
	"uc/pkg/email"
	"uc/pkg/logger"
	"uc/pkg/nacos"
)

func SendEmailStart() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Errorf("SendEmailStart painc: %v", r)
		}
	}()
	conn, err := AMQP.Get()
	if err != nil {
		logger.Logger.Errorf("SendEmailStart AMQP.Get error:%v", err)
		panic(err)
	}
	defer AMQP.Put(conn)

	ch, err := conn.Channel()
	if err != nil {
		logger.Logger.Errorf("SendEmailStart conn.Channel error:%v", err)
		panic(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(nacos.Config.RabbitMq.Queues.SendEmail, "", false, false, false, false, nil)
	if err != nil {
		logger.Logger.Errorf("SendEmailStart ch.Consume error:%v", err)
		panic(err)
	}
	for {
		select {
		case msg := <-msgs:
			logger.Logger.Infof("SendEmailStart sendEmail get:%v", string(msg.Body))
			err = sendEmail(msg.Body)
			// 错误有可能是邮箱不存在，所以发送过的邮箱不再处理
			if err != nil {
				logger.Logger.Errorf("SendEmailStart sendEmail error:%v", err)
			}
			err = msg.Ack(true)
			if err != nil {
				logger.Logger.Errorf("SendEmailStart ack error:%v", err)
				return
			}
		}
	}

}

func sendEmail(data []byte) error {
	var emailSendData types.SendEmailData

	err := json.Unmarshal(data, &emailSendData)
	if err != nil {
		logger.Logger.Errorf("sendEmail json.Unmarshal error:%v", err)
		return err
	}

	emailOptions := constant.SendEmailOptionsData[emailSendData.Behavior]
	title := emailOptions.Title[emailSendData.Language]
	templateUrl := emailOptions.Template[emailSendData.Language]
	tmpl, err := template.ParseFiles(templateUrl)
	if err != nil {
		panic(err)
	}
	// 创建邮件内容
	emailData := map[string]interface{}{
		"Code": emailSendData.Data,
	}
	// 将数据填充到模板中
	var body bytes.Buffer
	err = tmpl.Execute(&body, emailData)
	if err != nil {
		logger.Logger.Errorf("sendEmail tmpl.Execute error:%v", err)
		return err
	}

	err = email.MyEmail.SendEmail(title, []string{emailSendData.Email}, email.MAIL_TYPE_HTML, body.String())
	if err != nil {
		return err
	}
	return nil
}
