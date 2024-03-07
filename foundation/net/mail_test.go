package net

import "testing"

func TestSendMail(t *testing.T) {
	mailServer := &ServerInfo{
		Server: "smtp.126.com",
		Port:   25,
	}
	mailSender := &SenderInfo{
		User:     "",
		Password: "",
	}

	mailInfo := &MailInfo{
		SendTo:     []string{"rangp", "rangm"},
		CCTo:       []string{"muim"},
		Subject:    "About SendMail",
		Content:    "SendMail Code",
		Attachment: []string{"/home/rangh/codespace/magicExam/exam13/main.go"},
	}

	err := SendMail(mailServer, mailSender, mailInfo)
	if err != nil {
		t.Errorf("SendMail failed, error:%s", err.Error())
	}
}
