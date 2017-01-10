package main

import (
	"github.com/scorredoira/email"
	"log"
	"net/mail"
	"net/smtp"
)

func sendMail(smtpServerAndPort, from, to, subject, body, chartPngPath string) {
	log.Printf("Sending email through %s...", smtpServerAndPort)

	m := email.NewMessage(subject, body)

	address, err := (&mail.AddressParser{}).Parse(from)
	if err != nil {
		log.Fatalf("Error from AddressParser.Parse('%s'): %s", from, err)
	}
	m.From = *address

	m.To = []string{to}

	if err := m.Attach(chartPngPath); err != nil {
		log.Fatalf("Error from m.Attach: %s", err)
	}

	c, err := smtp.Dial(smtpServerAndPort)
	if err != nil {
		log.Fatalf("Error from smtp.Dial('%s'): %s", smtpServerAndPort, err)
	}
	if err = c.Mail(from); err != nil {
		log.Fatalf("Error from c.Mail('%s'): %s", from, err)
	}
	if err = c.Rcpt(to); err != nil {
		log.Fatalf("Error from c.Rcpt('%s'): %s", to, err)
	}
	w, err := c.Data()
	if err != nil {
		log.Fatalf("Error from c.Data(): %s", err)
	}
	_, err = w.Write([]byte(m.Bytes()))
	if err != nil {
		log.Fatalf("Error from w.Write(msg): %s", err)
	}
	err = w.Close()
	if err != nil {
		log.Fatalf("Error from w.Close(): %s", err)
	}
	err = c.Quit()
	if err != nil {
		log.Fatalf("Error from c.Quit(): %s", err)
	}
	log.Printf("Email sent.")
}
