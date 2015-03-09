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

// The mode determines the maximum output encoding, 7Bit, 8Bit or Binary.
// See also MIME.WriteTo()
type Mode int

const (
	Mode7Bit Mode = iota
	Mode8Bit
	ModeBinary
)

const lineEnd string = "\r\n"
const lineShouldLength = 78
const lineMaxLength = 998
const headerBufSize = 50 * lineShouldLength
const headerBufSizeMime = 4 * lineShouldLength

var (
	ErrFromRequired    = errors.New("email: From is required")
	ErrInvalidMimeTree = errors.New("email: Ambigious MIME tree for inserting text or attachment")
	ErrNoBody          = errors.New("email: body is missing")
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
	Message MIME
}

type MIME interface {
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
	buf := new(bytes.Buffer)
	err := e.WriteTo(buf, m)
	return buf.Bytes(), err
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

func writeEscapeHeader(b *bytes.Buffer, key, value string) {
	line := []byte(fmt.Sprintf("%s: ", textproto.CanonicalMIMEHeaderKey(key)))
	for _, word := range strings.SplitAfter(value, " ") {
		esc := escapeWord(word)
		if len(line)+len(esc) > lineShouldLength {
			b.Write(line)
			b.WriteString(lineEnd)
			line = append([]byte(" "), esc...)
		} else {
			line = append(line, esc...)
		}
	}
	b.Write(line)
	b.WriteString(lineEnd)
}

func writeEscapeAddressHeader(b *bytes.Buffer, key string, addresses ...mail.Address) {
	line := []byte(fmt.Sprintf("%s:", textproto.CanonicalMIMEHeaderKey(key)))

	for i, address := range addresses {
		if i != 0 {
			line = append(line, ',')
		}

		s := address.String()
		if len(line)+len(s) >= lineShouldLength {
			line = append(line, lineEnd...)
			b.Write(line)
			line = []byte{}
		}
		line = append(line, ' ')
		line = append(line, s...)
	}

	line = append(line, lineEnd...)
	b.Write(line)
}

func writeBoundary(w io.Writer, boundary string) error {
	if _, err := w.Write([]byte(lineEnd + boundary)); err != nil {
		return err
	}
	return nil
}

// Write this email to a writer. The mode determines what is the most liberal
// encoding the connection accepts. E.g. in case of 8-bit, binary objects will
// be base64-encoded and in case of 7-bit, utf8-text will be encoded as
// quoted-printable
func (e *Email) WriteTo(w io.Writer, m Mode) error {
	buf := &bytes.Buffer{}
	buf.Grow(headerBufSize)

	if e.Date.IsZero() {
		e.Date = time.Now()
	}
	if e.MessageId == "" {
		e.MessageId = genenerateMessageId()
	}
	if e.From.Address == "" {
		return ErrFromRequired
	}
	if e.Message == nil {
		return ErrNoBody
	}
	if e.Headers == nil {
		e.Headers = make(textproto.MIMEHeader)
	}

	writeEscapeHeader(buf, "Date", e.Date.Format(time.RFC1123Z))
	writeEscapeAddressHeader(buf, "From", e.From)

	if len(e.To) > 0 {
		writeEscapeAddressHeader(buf, "To", e.To...)
	}
	if len(e.Cc) > 0 {
		writeEscapeAddressHeader(buf, "Cc", e.Cc...)
	}
	if len(e.Bcc) > 0 {
		writeEscapeAddressHeader(buf, "Bcc", e.Bcc...)
	}
	if len(e.ReplyTo) > 0 {
		writeEscapeAddressHeader(buf, "Reply-To", e.ReplyTo...)
	}
	writeEscapeHeader(buf, "Subject", e.Subject)
	writeEscapeHeader(buf, "Message-Id", e.MessageId)

	e.Headers.Add("MIME-Version", "1.0")

	for key, values := range e.Headers {
		for _, value := range values {
			writeEscapeHeader(buf, key, value)
		}
	}

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}
	return e.Message.WriteTo(w, m)
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
	textPart := &MIMEPart{
		Type:        "text/plain",
		Disposition: "inline",
		Charset:     "utf-8",
		Headers:     textproto.MIMEHeader{},
		Content:     r,
	}

	switch p := e.Message.(type) {
	case nil:
		e.Message = textPart
	case *MIMEMultipart:
		p.Parts = append(p.Parts, textPart)
	case *MIMEPart:
		e.Message = &MIMEMultipart{
			Type:  "multipart/alternative",
			Parts: []MIME{p, textPart},
		}
	default:
		return ErrInvalidMimeTree
	}
	return nil
}

// Add a text body to this message. The text must be UTF-8. Adding multiple text
// bodies is not recommended, but will not throw an error.
func (e *Email) AddTextBodyString(s string) error {
	return e.AddTextBody(bytes.NewBufferString(s))
}

func (e *Email) AddHTMLBody(r io.Reader, charset string) error {
	htmlPart := &MIMEPart{
		Type:        "text/html",
		Disposition: "inline",
		Charset:     "utf-8",
		Headers:     textproto.MIMEHeader{},
		Content:     r,
	}

	switch p := e.Message.(type) {
	case nil:
		e.Message = htmlPart
	case *MIMEMultipart:
		p.Parts = append(p.Parts, htmlPart)
	case *MIMEPart:
		e.Message = &MIMEMultipart{
			Type:  "multipart/alternative",
			Parts: []MIME{htmlPart, p},
		}
	default:
		return ErrInvalidMimeTree
	}
	return nil
}

type MIMEPart struct {
	Type        string
	Disposition string
	Charset     string
	Headers     textproto.MIMEHeader
	Content     io.Reader
}

func (p *MIMEPart) Bytes() ([]byte, error) {
	return nil, nil
}

func (p *MIMEPart) WriteTo(w io.Writer, m Mode) error {
	buf := &bytes.Buffer{}
	buf.Grow(headerBufSizeMime)

	writeEscapeHeader(buf, "Content-Type", p.Type)

	buf.WriteString(lineEnd)

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}
	_, err := io.Copy(w, p.Content)
	return err
}

type MIMEMultipart struct {
	Type     string
	Headers  textproto.MIMEHeader
	Boundary string
	Parts    []MIME
}

func generateBoundary() string {
	return "boundary"
}

func genenerateMessageId() string {
	return "message id"
}

func generateContentId() string {
	return "content id"
}

func (p *MIMEMultipart) WriteTo(w io.Writer, m Mode) error {
	buf := &bytes.Buffer{}
	buf.Grow(headerBufSizeMime)

	if p.Boundary == "" {
		p.Boundary = generateBoundary()
	}

	writeEscapeHeader(buf, "Content-Type", p.Type)

	buf.WriteString(lineEnd)

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}

	if err := writeBoundary(w, p.Boundary); err != nil {
		return err
	}

	for _, part := range p.Parts {
		if err := part.WriteTo(w, m); err != nil {
			return err
		} else if err = writeBoundary(w, p.Boundary); err != nil {
			return err
		}
	}
	return nil
}
