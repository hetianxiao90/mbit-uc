package rabbitmq

import (
	"bytes"
	"encoding/json"
	"html/template"
	"uc/configs"
	"uc/internal/constant"
	"uc/internal/types"
	"uc/pkg/email"
	"uc/pkg/logger"
)

func SendEmailStart() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("SendEmailStart painc: %v", r)
		}
	}()
	conn, err := AMQP.Get()
	if err != nil {
		logger.Errorf("SendEmailStart AMQP.Get error:%v", err)
		panic(err)
	}
	defer AMQP.Put(conn)

	ch, err := conn.Channel()
	if err != nil {
		logger.Errorf("SendEmailStart conn.Channel error:%v", err)
		panic(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(configs.Config.RabbitMq.Queues.SendEmail, "", false, false, false, false, nil)
	if err != nil {
		logger.Errorf("SendEmailStart ch.Consume error:%v", err)
		panic(err)
	}
	for {
		select {
		case msg := <-msgs:
			logger.Infof("SendEmailStart sendEmail get:%v", string(msg.Body))
			err = sendEmail(msg.Body)
			if err != nil {
				logger.Errorf("SendEmailStart sendEmail error:%v", err)
				return
			}
			err = msg.Ack(true)
			if err != nil {
				logger.Errorf("SendEmailStart ack error:%v", err)
				return
			}
		}
	}

}

func sendEmail(data []byte) error {
	var emailSendData types.SendEmailData

	err := json.Unmarshal(data, &emailSendData)
	if err != nil {
		logger.Errorf("sendEmail json.Unmarshal error:%v", err)
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
		logger.Errorf("sendEmail tmpl.Execute error:%v", err)
		return err
	}

	err = email.MyEmail.SendEmail(title, []string{emailSendData.Email}, email.MAIL_TYPE_HTML, body.String())
	if err != nil {
		logger.Errorf("sendEmail email.MyEmail.SendEmail error:%v", err)
		return err
	}
	return nil
}
