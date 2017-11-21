package uuid

import (
	"crypto/rand"
	"fmt"
)

// New generates random UUID according to RFC 4122
func New() (uuid string, err error) {
	raw := [16]byte{}
	_, err = rand.Read(raw[:16])
	if err != nil {
		return ``, err
	}

	raw[8] = raw[8]&^0xc0 | 0x80
	raw[6] = raw[6]&^0xf0 | 0x40
	uuid = fmt.Sprintf(`%x-%x-%x-%x-%x`, uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	return
}
