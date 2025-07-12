//go:build linux

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

// Returns the SHA-256 hash of the data stored in /etc/machine-id.
func linuxGetKey() []byte {
	mId := readFile("/etc/machine-id")

	h := sha256.New()
	h.Write(mId)

	zero(&mId)

	return h.Sum(nil)
}

// Encrypts some data and returns a pointer pointing to the encrypted byte slice.
// Also zeros the given data.
// Note: On Windows, this function will benefit from Windows's [Data Protection API (DPAPI)]
// by using [billgraziano]'s [Go wrapper]. On Linux, this function will utilize the
// symmetric [AES-GCM] cipher. The AES key will be the SHA-256 hash of the data stored in
// /etc/machine-id. This is implemented using the Go packages [crypto/aes] and [crypto/cipher].
//
// [Data Protection API (DPAPI)]: https://wikipedia.org/wiki/Data_Protection_API
// [billgraziano]: https://github.com/billgraziano
// [Go wrapper]: https://pkg.go.dev/github.com/billgraziano/dpapi
// [AES-GCM]: https://wikipedia.org/wiki/Galois/Counter_Mode
func encrypt(data []byte) *[]byte {
	key := linuxGetKey()

	cipherBlock, err := aes.NewCipher(key)
	checkError(err)

	gcm, err := cipher.NewGCM(cipherBlock)
	checkError(err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	checkError(err)

	encrypted := gcm.Seal(nonce, nonce, data, nil)

	zero(&key)
	zero(&data)
	zero(&nonce)

	return &encrypted
}

// Decrypts some data and returns a pointer pointing to the decrypted byte slice.
// Also zeros the given data.
// Note: On Windows, this function will benefit from Windows's [Data Protection API (DPAPI)]
// by using [billgraziano]'s [Go wrapper]. On Linux, this function will utilize the
// symmetric [AES-GCM] cipher. The AES key will be the SHA-256 hash of the data stored in
// /etc/machine-id. This is implemented using the Go packages [crypto/aes] and [crypto/cipher].
//
// [Data Protection API (DPAPI)]: https://wikipedia.org/wiki/Data_Protection_API
// [billgraziano]: https://github.com/billgraziano
// [Go wrapper]: https://pkg.go.dev/github.com/billgraziano/dpapi
// [AES-GCM]: https://wikipedia.org/wiki/Galois/Counter_Mode
func decrypt(data []byte) *[]byte {
	key := linuxGetKey()

	cipherBlock, err := aes.NewCipher(key)
	checkError(err)

	gcm, err := cipher.NewGCM(cipherBlock)
	checkError(err)

	nonceSize := gcm.NonceSize()

	if len(data) < nonceSize {
		panic(errors.New("Encrypted data size is smaller than the required nonceSize"))
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, encrypted, nil)
	checkError(err)

	zero(&key)
	zero(&data)
	zero(&nonce)
	zero(&encrypted)

	return &decrypted
}
