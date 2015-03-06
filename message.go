/*
Package email is for the construction of email messages.

The defaults of this package are to create RFC5322 compliant messages that are
accepted and usable by the majority of existing, popular email clients.
*/
package email

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

// An email contains all field needing for constructing the message. The actual
// message is a tree of email (MIME) parts.
type Email struct {
	From                 mail.Address
	To, Cc, Bcc, ReplyTo []mail.Address

	// Defaults to Now
	Date    time.Time
	Subject string

	// This will be auto generated if not provided
	MessageId string

	// Optional headers
	Headers textproto.MIMEHeader

	// Actual message
	Message MIMEPart
}

// The mode determines the maximum output encoding, 7Bit, 8Bit or Binary.
// See also Message.WriteTo()
type Mode int

const (
	Mode7Bit Mode = iota
	Mode8Bit
	ModeBinary
)

const lineEnd string = "\r\n"
const lineShouldLength = 78
const lineMaxLength = 998

type Message interface {
	WriteTo(w io.Writer, m Mode) error
}

// Create a new email with the required and most important headers filled.
func New(subject string, from mail.Address, to ...mail.Address) *Email {
	return &Email{Subject: subject, From: from, To: to}
}

// Add To-receipient
func (e *Email) AddTo(a ...mail.Address) {
	e.To = append(e.To, a...)
}

// Add CC-receipient
func (e *Email) AddCc(a ...mail.Address) {
	e.Cc = append(e.Cc, a...)
}

// Add BCC-receipient
func (e *Email) AddBcc(a ...mail.Address) {
	e.Bcc = append(e.Bcc, a...)
}

// Export this emails to a byte string. It calls ToWriter.
func (e *Email) Bytes(m Mode) ([]byte, error) {
	b := new(bytes.Buffer)
	err := e.WriteTo(b, m)
	return b.Bytes(), err
}

func escapeWord(word string) []byte {
	needsEscape := false
	for _, r := range word {
		if r < ' ' || r > '~' {
			needsEscape = true
			break
		}
	}

	if !needsEscape {
		return []byte(word)
	}

	result := []byte{}
	line := bytes.NewBufferString("=?utf-8?q?")

	for i := 0; i < len(word); i++ {

		if line.Len()+5 > lineShouldLength {
			result = append(result, line.Bytes()...)
			result = append(result, "?="...)
			result = append(result, lineEnd...)
			line = bytes.NewBufferString(" =?utf-8?q?")
		}
		switch c := word[i]; {
		case c == ' ':
			line.WriteByte('_')
		case c >= '!' && c <= '~' && c != '=' && c != '?' && c != '_':
			line.WriteByte(c)
		default:
			fmt.Fprintf(line, "=%02X", c)
		}
	}

	result = append(result, line.Bytes()...)
	result = append(result, "?="...)
	return result

}

func writeEscapeHeader(w io.Writer, key, value string) {
	line := []byte(fmt.Sprintf("%s: ", textproto.CanonicalMIMEHeaderKey(key)))
	for _, word := range strings.SplitAfter(value, " ") {
		esc := escapeWord(word)
		if len(line)+len(esc) > lineShouldLength {
			w.Write(line)
			w.Write([]byte(lineEnd))
			line = append([]byte(" "), esc...)
		} else {
			line = append(line, esc...)
		}
	}
	w.Write(line)
	w.Write([]byte(lineEnd))
}

func writeEscapeAddressHeader(w io.Writer, key string, addresses ...mail.Address) {
	line := []byte(fmt.Sprintf("%s:", textproto.CanonicalMIMEHeaderKey(key)))

	for i, address := range addresses {
		if i != 0 {
			line = append(line, ',')
		}

		s := address.String()
		if len(line)+len(s) >= lineShouldLength {
			line = append(line, lineEnd...)
			w.Write(line)
			line = []byte{}
		}
		line = append(line, ' ')
		line = append(line, s...)
	}

	line = append(line, lineEnd...)
	w.Write(line)
}

// Write this email to a writer. The mode determines what is the most liberal
// encoding the connection accepts. E.g. in case of 8-bit, binary objects will
// be base64-encoded and in case of 7-bit, utf8-text will be encoded as
// quoted-printable
func (e *Email) WriteTo(w io.Writer, m Mode) error {
	if e.Date.IsZero() {
		e.Date = time.Now()
	}
	if e.MessageId == "" {
		e.MessageId = genMessageId()
	}
	if e.From.Address == "" {
		return errors.New("email: From is required")
	}

	writeEscapeHeader(w, "Date", e.Date.Format(time.RFC1123Z))
	writeEscapeAddressHeader(w, "From", e.From)
	if len(e.To) > 0 {
		writeEscapeAddressHeader(w, "To", e.To...)
	}
	if len(e.Cc) > 0 {
		writeEscapeAddressHeader(w, "Cc", e.Cc...)
	}
	if len(e.Bcc) > 0 {
		writeEscapeAddressHeader(w, "Bcc", e.Bcc...)
	}
	if len(e.ReplyTo) > 0 {
		writeEscapeAddressHeader(w, "Reply-To", e.ReplyTo...)
	}

	writeEscapeHeader(w, "Subject", e.Subject)
	writeEscapeHeader(w, "Message-Id", e.MessageId)

	if e.Headers == nil {
		e.Headers = make(textproto.MIMEHeader)
	}
	for key, values := range e.Headers {
		for _, value := range values {
			writeEscapeHeader(w, key, value)
		}
	}

	w.Write([]byte(lineEnd))
	w.Write([]byte("body"))
	w.Write([]byte(lineEnd))
	return nil
}

// Add a header to the message. These headers are not validated, and headers
// that are represented in another field are throwing an error.
func (e *Email) AddHeader(key, value string) error {
	if e.Headers == nil {
		e.Headers = make(textproto.MIMEHeader)
	}
	e.Headers.Add(key, value)
	return nil
}

// Add a text body to this message. The text must be UTF-8. Adding multiple text
// bodies is not recommended, but will not throw an error.
func (e *Email) AddTextBody(r io.Reader) error {
	//if e.Message == nil {
	//	e.Message = &MIMEPart{
	//		Type:         "text/plain",
	//		Disposition:  "inline",
	//		Charset:      charset,
	//		ExtraHeaders: textproto.MIMEHeader{},
	//		Content:      r,
	//	}
	//}
	return nil
}

// Add a text body to this message. The text must be UTF-8. Adding multiple text
// bodies is not recommended, but will not throw an error.
func (e *Email) AddTextBodyString() error {
	return nil
}

func (e *Email) AddHTMLBody(r io.Reader, charset string) error {
	return nil
}

func genMessageId() string {
	return "message id"
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
