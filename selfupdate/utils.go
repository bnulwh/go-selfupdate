package selfupdate

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

func GenerateSha256(path string) string {
	h := sha256.New()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	h.Write(b)
	sum := h.Sum(nil)
	//return sum
	return base64.URLEncoding.EncodeToString(sum)
}
