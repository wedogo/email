package email

import (
	"bytes"
	"io"
	"net/mail"
	"net/textproto"
	"time"
)

type Message interface {
	Headers() textproto.MIMEHeader
	Bytes() ([]byte, error)
	WriteTo(w io.Writer, m Mode) error
}

type MIMEPart struct {
	Type         string
	Disposition  string
	Charset      string
	ExtraHeaders textproto.MIMEHeader
	Content      io.Reader
}

func (p *MIMEPart) Headers() textproto.MIMEHeader {
	return nil
}

func (p *MIMEPart) Bytes() ([]byte, error) {
	return nil, nil
}

func (p *MIMEPart) WriteTo(w io.Writer, m Mode) error {
	return nil
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
	return &Email{Subject: subject, From: []mail.Address{from}, To: to}
}

// A message represents
type Email struct {
	Sender                     mail.Address
	From, To, Cc, Bcc, ReplyTo []mail.Address

	Date    time.Time
	Subject string

	MessageId string

	// Optional headers
	Headers textproto.MIMEHeader

	// Actual message
	Message Message
}

func (e *Email) AddTo(a ...mail.Address) {
	e.To = append(e.To, a...)
}

func (e *Email) AddCc(a ...mail.Address) {
	e.Cc = append(e.Cc, a...)
}

func (e *Email) AddBcc(a ...mail.Address) {
	e.Bcc = append(e.Bcc, a...)
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

func (e *Email) AddHeader(key, value string) error {
	e.Headers.Add(key, value)
	return nil
}

func (e *Email) AddTextBody(r io.Reader, charset string) error {
	if e.Message == nil {
		e.Message = &MIMEPart{
			Type:         "text/plain",
			Disposition:  "inline",
			Charset:      charset,
			ExtraHeaders: textproto.MIMEHeader{},
			Content:      r,
		}
	}
	return nil
}

func (e *Email) AddHTMLBody(r io.Reader, charset string) error {
	return nil
}

func (e *Email) AddAttachment(r io.Reader, filename, MIMEType string) error {
	return nil
}

func (e *Email) AddTextAttachment(r io.Reader, filename, MIMEType, charset string) error {
	return nil
}

func genMessageId() string {
	return "message id"
}
