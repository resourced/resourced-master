// Package libsmtp provides SMTP related library functions.
package libsmtp

import (
	"encoding/base64"
	"fmt"
)

// BuildMessage is a helper function to build email message.
func BuildMessage(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["Return-Path"] = from
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
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
