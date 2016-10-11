package goemail

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"
)

type Server struct {
	SmtpServer     string
	SmtpPort       int
	EmailSender    string
	PasswordSender string
}

func NewServer(smtpServer string, smtpPort int, emailSender, passwordSender string) *Server {
	return &Server{SmtpServer: smtpServer, SmtpPort: smtpPort, EmailSender: emailSender, PasswordSender: passwordSender}
}

func (s *Server) SendEmail(to, subject, body string) error {
	toEmail, _ := mail.ParseAddress(to)

	auth := smtp.PlainAuth("", s.EmailSender, s.PasswordSender, s.SmtpServer)

	header := make(map[string]string)
	header["Return-Path"] = s.EmailSender
	header["From"] = s.EmailSender
	header["To"] = toEmail.String()
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	err := sendMailInt(s.SmtpServer+":"+strconv.Itoa(s.SmtpPort),
		auth,
		s.EmailSender,
		[]string{to},
		[]byte(message),
		s.SmtpServer,
	)

	return err
}

func sendMailInt(addr string, a smtp.Auth, from string, to []string, msg []byte, mailserver string) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: mailserver, InsecureSkipVerify: true}
		//config := &tls.Config{ServerName: c.serverName}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}

	if a != nil {
		ok, _ := c.Extension("AUTH")
		if ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
