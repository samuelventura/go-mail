package mail

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"sort"
	"strings"
	"time"
)

//with display name
//User Name <user.name@domain.tld>
//dig google.com MX

func MailSend(args Args) error {
	id := args.Get("id").(string)
	from := args.Get("from").(string)
	to := args.Get("to").(string)
	subject := args.Get("subject").(string)
	mime := args.Get("mime").(string)
	body := args.Get("body").([]byte)
	pkey := args.Get("pkey").([]byte)
	afrom, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}
	ato, err := mail.ParseAddress(to)
	if err != nil {
		return err
	}
	toDomain := strings.Split(ato.Address, "@")[1]
	fromDomain := strings.Split(afrom.Address, "@")[1]
	msg, bodyLen := mailPack(id, ato.String(), afrom.String(), subject, mime, body)
	err = dkimSign(&msg, bodyLen, fromDomain, []byte(pkey))
	if err != nil {
		return err
	}
	mxs, err := net.LookupMX(toDomain)
	if err != nil {
		return err
	}
	sort.Slice(mxs, func(i, j int) bool {
		return mxs[i].Pref < mxs[j].Pref
	})
	mxsn := make([]string, 0, len(mxs))
	for _, x := range mxs {
		mxsn = append(mxsn, x.Host)
		addr := fmt.Sprintf("%s:25", x.Host)
		args.Set("addr", addr)
		args.Set("from", afrom.Address)
		args.Set("to", []string{ato.Address})
		args.Set("msg", msg)
		err = smtpSend(args)
		_, ok := err.(*DialError)
		if ok {
			continue
		}
		return err
	}
	return fmt.Errorf("no working mx %v", mxsn)
}

func escapeHeader(str string) string {
	return mime.QEncoding.Encode("utf-8", str)
}

func mailPack(id string, to string, from string, subject string, mime string, body []byte) ([]byte, uint) {
	var b strings.Builder
	//Date: Mon, 04 Oct 2021 21:46:06 +0000
	const RFC822 = "Mon, 02 Jan 2006 15:04:05 -0700"
	fmt.Fprintf(&b, "%s: %s\r\n", "Message-Id", id)
	fmt.Fprintf(&b, "%s: %s\r\n", "Date", time.Now().Format(RFC822))
	fmt.Fprintf(&b, "%s: %s\r\n", "From", from)
	fmt.Fprintf(&b, "%s: %s\r\n", "To", to)
	fmt.Fprintf(&b, "%s: %s\r\n", "Subject", escapeHeader(subject))
	fmt.Fprintf(&b, "%s: %s\r\n", "MIME-Version", "1.0")
	fmt.Fprintf(&b, "%s: %s\r\n", "Content-Type", fmt.Sprintf("%s; charset=\"utf-8\"", mime))
	fmt.Fprintf(&b, "%s: %s\r\n", "Content-Transfer-Encoding", "base64")
	b.WriteString("\r\n")
	b.WriteString(base64.StdEncoding.EncodeToString(body))
	return []byte(b.String()), uint(len(body))
}
