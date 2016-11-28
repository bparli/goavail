package encrypt

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_EncryptDecrypt(t *testing.T) {
	Convey("Init the Global Map", t, func() {
		originalText := "Test Some Dummy String"
		key := []byte("example key 1234")
		cryptoText := Encrypt(key, originalText)
		plainText := Decrypt(key, cryptoText)
		So(plainText, ShouldEqual, "Test Some Dummy String")
	})
}
