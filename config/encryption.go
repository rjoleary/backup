package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

const (
	lenMagic  = 32
	lenSalt   = 8
	lenNonce  = 24
	lenHeader = lenMagic + lenSalt + lenNonce

	minPasswordLen = 6
	lenKey         = 32
	magicString    = "backuprc 0.1"
)

var magicWithPadding = append([]byte(magicString), make([]byte, lenMagic-len(magicString))...)

func genKey(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, 1<<15, 8, 1, lenKey)
}

func decrypt(file []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password is empty")
	}

	if len(file) < lenHeader {
		return nil, errors.New("header is too small")
	}

	magic, file := file[:lenMagic], file[lenMagic:]
	salt, file := file[:lenSalt], file[lenSalt:]
	nonce, file := file[:lenNonce], file[lenNonce:]
	cipherText := file

	if string(magic) != string(magicWithPadding) {
		return nil, errors.New("invalid magic")
	}

	key, err := genKey(password, salt)
	if err != nil {
		return nil, err
	}

	plainText, ok := secretbox.Open(nil, cipherText, (*[lenNonce]byte)(nonce), (*[lenKey]byte)(key))
	if !ok {
		return nil, errors.New("decryption error")
	}
	return plainText, nil
}

func encrypt(plainText []byte, password string) ([]byte, error) {
	if len(password) < minPasswordLen {
		return nil, fmt.Errorf("password is too short (min %d chars)", minPasswordLen)
	}

	extendLen := func(buf []byte, size int) (head, tail []byte) {
		return buf[:len(buf)+size], buf[len(buf) : len(buf)+size]
	}

	var buf = make([]byte, 0, lenHeader+secretbox.Overhead+len(plainText))
	buf, magic := extendLen(buf, lenMagic)
	buf, saltAndNonce := extendLen(buf, lenSalt+lenNonce)
	salt, nonce := saltAndNonce[0:lenSalt], saltAndNonce[lenSalt:]

	copy(magic, magicWithPadding)

	// Generate salt and nonce in a single call for simplicity.
	log.Println("Generating randomness...")
	if _, err := io.ReadFull(rand.Reader, saltAndNonce); err != nil {
		return nil, fmt.Errorf("error generating randomness: %v", err)
	}
	key, err := genKey(password, salt)
	if err != nil {
		return nil, err
	}

	return secretbox.Seal(buf, plainText, (*[lenNonce]byte)(nonce), (*[lenKey]byte)(key)), nil
}
