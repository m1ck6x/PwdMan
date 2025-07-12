package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const FILE_MODE os.FileMode = 0640

// A struct containing information about a specific user account.
type account struct {
	Service     string `json:"service"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	User        string `json:"user"`
	Pw          string `json:"pw"`
}

// Returns a pointer to a slice of structs containing information about a specific account.
// If the accounts file doesn't exist or is empty, then this function will return a pointer
// to an empty slice of above mentioned structs.
func getAllAccounts() *[]account {
	accounts := []account{}

	encryptedData := readFileRel("accounts")
	if len(encryptedData) == 0 {
		return &accounts
	}

	data := decrypt(encryptedData)

	json.Unmarshal(*data, &accounts)

	zero(data)
	zero(&encryptedData)

	return &accounts
}

// Takes in a pointer to a slice of account structs, encrypts the data in JSON format
// and then writes the resulting bytes into the local, predefined accounts file.
func saveAccountsToDisk(accounts *[]account) {
	data, err := json.Marshal(*accounts)
	checkError(err)

	encryptedData := encrypt(data)

	WriteFileRel("accounts", *encryptedData)

	zero(encryptedData)
	zero(&data)

	runtime.GC()
}

// Checks whether the supplied error is nil. If not, this function panics the error.
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// Takes a pointer to a slice of bytes and sets every byte within that slice to zero.
func zero(slice *[]byte) {
	for i := range *slice {
		(*slice)[i] = 0
	}

	runtime.KeepAlive(*slice)
}

// This function checks whether a password is "valid". A password is considered "valid"
// if its SHA-1 hash is not included in the data returned by querying [HIBPs' password search by range]
// api and there was no other error when querying (e.g. no internet connection, error in the hashing
// function, etc).
//
// Two values will be returned:
//   - a boolean indicating whether the password is considered "valid"
//   - a number indicating how many times the hash has appeared in the data set (refer to the
//     afore mentioned [HIBPs' password search by range] docs for more information). Note that this
//     number will be 0 if there was an error OR if the hash of the password just never appeared
//     in a data set and is therefor considered "valid". Hence, if the returned boolean is false
//     and the returned number == 0, an error must have occured.
//
// [HIBPs' password search by range]: https://haveibeenpwned.com/API/v3#SearchingPwnedPasswordsByRange
func isPwValid(pw []byte) (bool, uint32) {
	// deepcode ignore InsecureHash: The pwned passwords API only supports SHA-1 or NTLM hashes. I went with SHA-1. The actual hash is only used for querying the pwned passwords API, relativizing this issue.
	h := sha1.New()
	_, err := h.Write(pw)

	defer h.Reset()

	if err != nil {
		return false, 0
	}

	finalHash := strings.TrimSpace(strings.ToUpper(hex.EncodeToString(h.Sum(nil))))
	prefix, suffix := finalHash[:5], finalHash[5:]

	resp, err := http.Get(fmt.Sprint("https://api.pwnedpasswords.com/range/", prefix))

	if err != nil || resp.StatusCode != 200 {
		return false, 0
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return false, 0
	}

	entries := strings.SplitSeq(string(data), "\r\n")

	for entry := range entries {
		split := strings.Split(entry, ":")
		pwdSuffix := split[0]
		count, _ := (strconv.Atoi(split[1]))

		if pwdSuffix == suffix {
			return false, uint32(count)
		}
	}

	return true, 0
}

// Generates a password and returns the pointer pointing to the password's byte slice.
// The utilized set of characters contains, in order:
//   - all 26 letters of the English alphabet in lower case
//   - all 26 letters of the English alphabet in upper case
//   - all 10 digits (0-9)
//   - the special characters ?, +, =, !, &, /, -, _, <, > and |
//
// Note that the digits appear thrice and the special characters appear twice in the
// set of characters to increase their probability/rate of appearing in the resulting password.
//
// The resulting password will consist of at least 10 and at most 20 characters.
// The utilized RNG is the cryptographically secure RNG implemented in the [crypto/rand] package.
//
// Don't use this function directly as it will not check if the generated password
// can be considered "valid". Use the "generatePw()" function instead.
func generatePwInternal() *[]byte {
	chars := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012345678901234567890123456789?+=!&/-_<>|?+=!&/-_<>|")

	bi, err := rand.Int(rand.Reader, big.NewInt(int64(21)))
	checkError(err)

	buf := make([]byte, bi.Int64()+10)

	for i := range buf {
		bi, err = rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		checkError(err)

		buf[i] = chars[bi.Int64()]
	}

	return &buf
}

// Generates a password and returns the pointer pointing to its byte slice
// using the "generatePwInternal()" function until the resulting bytes
// are considered a valid password by the "isPwValid()" function. In case of an error,
// a pointer to an empty byte slice is returned.
func generatePw() *[]byte {
	pwd := generatePwInternal()
	b, n := isPwValid(*pwd)

	for !b && n != 0 {
		zero(pwd)
		pwd = generatePwInternal()
	}

	if !b && n == 0 {
		return &[]byte{}
	} else {
		return pwd
	}
}

// Returns the string returned by [os.Getwd]. Or panics if the returned error != nil.
func getWd() string {
	wd, err := os.Getwd()
	checkError(err)

	return wd
}

// Reads a file from an absolute path.
func readFile(name string) []byte {
	data, err := os.ReadFile(name)
	if os.IsNotExist(err) {
		return []byte{}
	}

	checkError(err)

	return data
}

// Reads a file from a relative path (using the "getWd()" function).
func readFileRel(name string) []byte {
	return readFile(getWd() + string(os.PathSeparator) + name)
}

// Writes to a file from a relative path (using the "getWd()" function) using the in this
// file (global.go) defined FILE_MODE and the perm parameter.
func WriteFileRel(name string, data []byte) error {
	err := os.WriteFile(getWd()+string(os.PathSeparator)+name, data, FILE_MODE)

	zero(&data)

	return err
}
