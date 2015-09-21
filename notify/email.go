package notify

import (
	"io"
	"mime/multipart"
	"strconv"

	"github.com/go-gomail/gomail"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Email struct {
	From        string
	To          string
	CC          string
	Title       string
	ContentType string
	SMTP        map[string]interface{}
}

func (self *Email) Notify(body string, attachments map[string][]*multipart.FileHeader) {

	m := gomail.NewMessage()
	m.SetHeader("From", self.From)

	m.SetHeader("To", strings.Split(self.To, ",")...)

	if self.CC != "" {
		m.SetHeader("CC", strings.Split(self.CC, ",")...)
	}

	m.SetHeader("Subject", self.Title)
	m.SetBody(self.ContentType, body)

	for _, tmp := range attachments {
		for _, attachment := range tmp {
			f, err := attachment.Open()
			if err != nil {
				log.Panic(err)
			}

			defer f.Close()
			m.Attach(attachment.Filename,
				gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := io.Copy(w, f)
					return err
				}),
			)
		}
	}

	port, err := strconv.Atoi(self.SMTP["port"].(string))

	if err != nil {
		log.Panic(err)
	}

	d := gomail.NewPlainDialer(
		self.SMTP["host"].(string),
		port,
		self.SMTP["user"].(string),
		self.SMTP["password"].(string),
	)

	//	log.Info(d)
	log.Info(body)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Panic(err)
	}

}
