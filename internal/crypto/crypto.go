package crypto

import (
	"crypto/md5"
	"encoding/base64"
	"io"
)

// Md5Hash returns the md5 hash of the given data
func Md5Hash(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	sum := h.Sum(nil)
	b64Sum := base64.StdEncoding.EncodeToString(sum)
	return b64Sum
}
