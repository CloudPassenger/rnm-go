package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"

	"github.com/CloudPassenger/rnm-go/infra/pool"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

var (
	size = 8192
)

type RealityCipher struct {
}

func Value(vals ...byte) (value int) {
	for i, val := range vals {
		value |= int(val) << ((len(vals) - i - 1) * 8)
	}
	return
}

func (conf *RealityCipher) Verify(data []byte, privKey []byte) ([]byte, bool) {
	n := len(data)
	buf := pool.Get(n)
	defer pool.Put(buf)
	copy(buf, data)
	if n <= recordHeaderLen {
		// fmt.Println("not tls handshake")
		return nil, false
	}
	// Reality detect clientHello length
	clientHelloLen := 0
	c2sSaved := make([]byte, 0, size)
	c2sSaved = append(c2sSaved, buf[:n]...)
	if clientHelloLen == 0 && len(c2sSaved) > recordHeaderLen {
		if recordType(c2sSaved[0]) != recordTypeHandshake || Value(c2sSaved[1:3]...) != VersionTLS10 || c2sSaved[5] != typeClientHello {
			fmt.Println("Wrong TLS record, send to fallback")
			return nil, false
		}
		clientHelloLen = recordHeaderLen + Value(c2sSaved[3:5]...)
		// fmt.Printf("clientHello len: %d\n", clientHelloLen)
	}
	if clientHelloLen > size { // too long
		// fmt.Println("clientHello too long, send to fallback")
		return nil, false
	}
	// parse client hello message
	clientHello := &clientHelloMsg{}
	ret := clientHello.unmarshal(buf[recordHeaderLen:clientHelloLen])
	if !ret {
		// fmt.Println("parse hello message return false")
		return nil, false
	}
	// fmt.Printf("%+v\n", clientHello)
	// Reality: get keyshare
	for _, keyShare := range clientHello.keyShares {
		if keyShare.group != X25519 || len(keyShare.data) != 32 {
			continue
		}
		authKey, err := curve25519.X25519(privKey, keyShare.data)
		// fmt.Printf("REALITY AuthKey[:16]: %v\tPrivKey: %T\n", authKey[:16], privKey[:16])
		if err != nil {
			// fmt.Println("authKey verification failed, aborted.")
			break
		}
		// No need for
		if _, err = hkdf.New(sha256.New, authKey, clientHello.random[:20], []byte("REALITY")).Read(authKey); err != nil {
			// fmt.Println("HKDF verification failed, aborted.")
			break
		}

		var aead cipher.AEAD
		if aesgcmPreferred(clientHello.cipherSuites) {
			block, _ := aes.NewCipher(authKey)
			aead, _ = cipher.NewGCM(block)
		} else {
			aead, _ = chacha20poly1305.New(authKey)
		}

		// fmt.Printf("REALITY AuthKey[:16]: %v\tAEAD: %T\n", authKey[:16], aead)
		ciphertext := make([]byte, 32)
		plainText := make([]byte, 32)
		copy(ciphertext, clientHello.sessionId)
		copy(clientHello.sessionId, plainText)
		// realRaw := bytes.Trim(clientHello.raw, "\x00")
		// fmt.Printf("%+v\n", realRaw)
		// copy(clientHello.sessionId, plainText) // hs.clientHello.sessionId points to hs.clientHello.raw[39:]
		if _, err = aead.Open(plainText[:0], clientHello.random[20:], ciphertext, clientHello.raw); err != nil {
			// fmt.Println(err)
			break
		}
		// copy(clientHello.sessionId, ciphertext)
		return data, true

	}
	// fmt.Errorf("no REALITY detected")
	return nil, false
}
