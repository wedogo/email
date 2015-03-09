package email

import (
	"bytes"
	"net/mail"
	"testing"
	"time"
)

func TestEscapeWord(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			"abcde",
			"abcde",
		},
		{
			"测试",
			"=?utf-8?q?=E6=B5=8B=E8=AF=95?=",
		},
		{
			"蒏慡戫溿煔煃肒芅邥瑐瑍覟晛桼桾粞絧絏蘹蠮褅褌諃aa蚔",
			"=?utf-8?q?=E8=92=8F=E6=85=A1=E6=88=AB=E6=BA=BF=E7=85=94=E7=85=83=E8=82=92=E8?=\r\n =?utf-8?q?=8A=85=E9=82=A5=E7=91=90=E7=91=8D=E8=A6=9F=E6=99=9B=E6=A1=BC=E6?=\r\n =?utf-8?q?=A1=BE=E7=B2=9E=E7=B5=A7=E7=B5=8F=E8=98=B9=E8=A0=AE=E8=A4=85=E8?=\r\n =?utf-8?q?=A4=8C=E8=AB=83aa=E8=9A=94?=",
		},
		{
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	for _, test := range tests {
		if o := escapeWord(test.input); string(o) != test.output {
			t.Errorf("escapeWord. Input %s, expected %s, got %s", test.input, test.output, string(o))
		}
	}
}

func TestWriteEscapeHeader(t *testing.T) {
	tests := []struct {
		k      string
		v      string
		output string
	}{
		{
			"message-id",
			"<jane@doe.com>",
			"Message-Id: <jane@doe.com>\r\n",
		},
		{
			"subject",
			"测试",
			"Subject: =?utf-8?q?=E6=B5=8B=E8=AF=95?=\r\n",
		},
		{
			"subjeCt",
			"蒏 慡戫 溿煔煃 肒芅邥 瑐瑍 覟 晛桼桾 粞絧絏 蘹蠮",
			"Subject: =?utf-8?q?=E8=92=8F_?==?utf-8?q?=E6=85=A1=E6=88=AB_?=\r\n =?utf-8?q?=E6=BA=BF=E7=85=94=E7=85=83_?=\r\n =?utf-8?q?=E8=82=92=E8=8A=85=E9=82=A5_?==?utf-8?q?=E7=91=90=E7=91=8D_?=\r\n =?utf-8?q?=E8=A6=9F_?==?utf-8?q?=E6=99=9B=E6=A1=BC=E6=A1=BE_?=\r\n =?utf-8?q?=E7=B2=9E=E7=B5=A7=E7=B5=8F_?==?utf-8?q?=E8=98=B9=E8=A0=AE?=\r\n",
		},
		{
			"subject",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"Subject: \r\n aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n",
		},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}

		if writeEscapeHeader(b, test.k, test.v); b.String() != test.output {
			t.Errorf("writeEscapeHeader. Input (%s,%s), expected %s, got %s", test.k, test.v, test.output, b.String())
		}
	}

}

func TestWriteEscapeAddressHeader(t *testing.T) {
	tests := []struct {
		k      string
		v      []mail.Address
		output string
	}{
		{
			"to",
			[]mail.Address{mail.Address{"测试", "test@example.com"}},
			"To: =?utf-8?q?=E6=B5=8B=E8=AF=95?= <test@example.com>\r\n",
		},
		{
			"BCC",
			[]mail.Address{mail.Address{"Very Long Name", "test@example.com"}, mail.Address{"Very Long Name", "test@example.com"}, mail.Address{"Very Long Name", "test@example.com"}, mail.Address{"Very Long Name", "test@example.com"}, mail.Address{"Very Long Name", "test@example.com"}},
			"Bcc: \"Very Long Name\" <test@example.com>, \"Very Long Name\" <test@example.com>,\r\n \"Very Long Name\" <test@example.com>, \"Very Long Name\" <test@example.com>,\r\n \"Very Long Name\" <test@example.com>\r\n",
		},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}

		if writeEscapeAddressHeader(b, test.k, test.v...); b.String() != test.output {
			t.Errorf("writeEscapeHeader. Input (%s,%s), expected %s, got %s", test.k, test.v, test.output, b.String())
		}
	}

}

func TestMessage(t *testing.T) {
	m := New("A test subject", mail.Address{"Test", "test@example.org"}, mail.Address{"To", "To@example.org"}, mail.Address{"To", "to2@example.org"})
	m.AddTo(mail.Address{"To3", "to3@example.org"})
	m.AddBcc(mail.Address{"BCC", "bcc@example.org"})
	m.AddCc(mail.Address{"CC1", "cc1@example.org"}, mail.Address{"CC2", "cc2@example.org"})
	m.ReplyTo = []mail.Address{mail.Address{"jane", "reply@example.org"}}
	m.Date, _ = time.Parse(time.RFC1123Z, time.RFC1123Z)
	m.MessageId = "<test@abc.org>"
	m.AddHeader("Foo", "Bar")
	m.AddHeader("Foo", "Bar")

	m.AddTextBodyString("Hello")
	b, err := m.Bytes(Mode8Bit)
	if err != nil {
		t.Error(err)
		return
	}

	mess, err := mail.ReadMessage(bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
		return
	}

	testHeaders := []struct {
		k      string
		output string
	}{
		{
			"From",
			"\"Test\" <test@example.org>",
		},
		{
			"Cc",
			"\"CC1\" <cc1@example.org>, \"CC2\" <cc2@example.org>",
		},
		{
			"Bcc",
			"\"BCC\" <bcc@example.org>",
		},
		{
			"To",
			"\"To\" <To@example.org>, \"To\" <to2@example.org>, \"To3\" <to3@example.org>",
		},
		{
			"Reply-To",
			"\"jane\" <reply@example.org>",
		},
		{
			"Subject",
			"A test subject",
		},
		{
			"Date",
			time.RFC1123Z,
		},
		{
			"Message-Id",
			"<test@abc.org>",
		},
		{
			"Foo",
			"Bar",
		},
	}

	for _, test := range testHeaders {
		if o := mess.Header.Get(test.k); o != test.output {
			t.Errorf("Message. Header %s, expected %s, got %s", test.k, test.output, o)
		}
	}

	if i := len(mess.Header["Foo"]); i != 2 {
		t.Errorf("Expected 2 items in Foo header, got %d", i)
	}

}

func TestMinimalMessage(t *testing.T) {
	m := Email{}
	if _, err := m.Bytes(Mode8Bit); err == nil {
		t.Errorf("Expected From error, but got %+v", err)
	}
	m.From = mail.Address{"Test", "test@example.org"}
	m.AddTextBodyString("")

	if b, err := m.Bytes(Mode8Bit); err != nil {
		t.Errorf("Unexpected error %+v", err)
		return
	} else {
		mess, err := mail.ReadMessage(bytes.NewBuffer(b))
		if err != nil {
			t.Error(err)
			return
		}

		headers := []string{"From", "Message-Id", "Date"}
		for _, h := range headers {
			if mess.Header.Get(h) == "" {
				t.Errorf("Header %s should always be present", h)
			}
		}

	}
}
