package mail

import "github.com/toorop/go-dkim"

func dkimSign(email *[]byte, bodyLength uint, domain string, privateKey []byte) error {
	options := dkim.NewSigOptions()
	options.PrivateKey = privateKey
	options.Domain = domain
	options.Selector = "dkim"
	options.SignatureExpireIn = 3600
	options.BodyLength = bodyLength
	options.Headers = []string{"message-id", "from", "to", "subject", "date", "mime-version", "content-type"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "relaxed/relaxed"
	return dkim.Sign(email, options)
}
