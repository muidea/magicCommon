package net

import (
	"crypto/tls"
	"gopkg.in/gomail.v2"
)

const (
	Html = "html"
	Text = "text"
)

// SendMail 发送邮件
// user : example@example.com login smtp server user
// password: xxxxx login smtp server password
// emailServer: smtp.example.com   smtp.163.com
// emailPort: 25
// sendTo, ccTo: example@example.com;example1@163.com;example2@sina.com.cn;...
// subject:The subject of mail
// content: The content of mail
// attachment: 附件
// mimeType: mail type html or text
// skipTls: 是否忽略Tls
func SendMail(user, password, emailServer string, emailPort int, sendTo, ccTo []string, subject, content string, attachment []string, mimeType string, skipTls bool) error {
	goMailMsg := gomail.NewMessage()
	goMailMsg.SetHeader("From", user)
	goMailMsg.SetHeader("To", sendTo...)
	if len(ccTo) > 0 {
		goMailMsg.SetHeader("Cc", ccTo...)
	}
	goMailMsg.SetHeader("Subject", subject)
	if mimeType == Html {
		goMailMsg.SetBody("text/html", content)
	} else {
		goMailMsg.SetBody("text/plain", content)
	}

	for _, val := range attachment {
		goMailMsg.Attach(val)
	}

	goMailDialer := gomail.NewDialer(emailServer, emailPort, user, password)
	if skipTls {
		goMailDialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return goMailDialer.DialAndSend(goMailMsg)
}
