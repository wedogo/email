package email

import (
	"bytes"
	"fmt"
	"net/mail"
	//"net/smtp"
	"testing"
)

func TestNormalMessage(t *testing.T) {
	m := New("A test subject 测试对象测试对象测试对象测试对a象测试对象测试对象 测试对象 测试对象测试对象 测试对象 测试对象", mail.Address{"测试", "from@example.org"}, mail.Address{"测试", "to@example.org"}, mail.Address{"测试", "to2@example.org"})
	b := &bytes.Buffer{}
	m.WriteTo(b, Mode8Bit)
	fmt.Println(b.String())

	mess, err := mail.ReadMessage(b)
	if err != nil {
		t.Error(err)
	}

	addrs, err := mess.Header.AddressList("To")

	if err != nil {
		t.Error(err)
	}

	for _, addr := range addrs {
		println(addr.Name)
	}
	if s := mess.Header.Get("Subject"); s != "A test subject 测试对象测试对象测试对象测试对象测试对象测试对象 测试对象 测试对象测试对象 测试对象 测试对象" {
		t.Errorf("Subject wrong, got: %s", s)
	}

}
