//go:build windows

package main

import (
	"github.com/billgraziano/dpapi"
)

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
	encrypted, err := dpapi.EncryptBytes(data)
	checkError(err)

	zero(&data)

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
	decrypted, err := dpapi.DecryptBytes(data)
	checkError(err)

	zero(&data)

	return &decrypted
}
