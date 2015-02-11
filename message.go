package email

import (
	"bytes"
	"io"
	"net/mail"
	"net/textproto"
	"time"
)

type Message interface {
	Headers() map[string]string
	Bytes() ([]byte, error)
	WriteTo(w io.Writer, m Mode) error
}

type MIMEPart struct {
	Type     string
	Encoding string
	Headers  map[string]string
	Content  io.Reader
}

type MIMEMultipart struct {
	MIMEPart
	Parts []MIMEPart
}

type Mode int

const (
	Mode7Bit Mode = iota
	Mode8Bit
	ModeBinary
)

func New(subject string, from mail.Address, to ...mail.Address) *Email {
	return &Email{Subject: subject, From: from, To: to}
}

// A message represents
type Email struct {
	From                 mail.Address
	To, Cc, Bcc, ReplyTo []mail.Address

	Date    time.Time
	Subject string

	MessageId string

	// Optional headers
	Headers textproto.MIMEHeader

	// Actual message
	Message Message
}

func (e *Email) AddTo(a ...mail.Address) {

}

func (e *Email) AddCc(a ...mail.Address) {

}

func (e *Email) AddBcc(a ...mail.Address) {

}

func (e *Email) Bytes(m Mode) ([]byte, error) {
	b := new(bytes.Buffer)
	err := e.ToWriter(b, m)
	return b.Bytes(), err
}

func (e *Email) ToWriter(w io.Writer, m Mode) error {
	if e.Date.IsZero() {
		e.Date = time.Now()
	}
	if e.MessageId == "" {
		e.MessageId = genMessageId()
	}

	return nil
}

func (m *Email) AddHeader(key, value string) error {
	return nil
}

func (e *Email) AddTextBody(r io.Reader, encoding string) error {
	return nil
}

func (e *Email) AddHTMLBody(r io.Reader, encoding string) error {
	return nil
}

func (e *Email) AddAttachment(r io.Reader, filename, MIMEType string) error {
	return nil
}

func (e *Email) AddTextAttachment(r io.Reader, filename, MIMEType, encoding string) error {
	return nil
}

func genMessageId() string {
	return "message id"
}
