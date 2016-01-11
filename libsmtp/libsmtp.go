package libsmtp

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
)

func EncodeRFC2047(str string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{str, ""}
	return strings.Trim(addr.String(), " <>")
}

func BuildMessage(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["Return-Path"] = from
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = EncodeRFC2047(subject)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = `text/plain; charset="utf-8"`
	headers["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	return message
}
