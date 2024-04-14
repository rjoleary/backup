package config

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	for _, tt := range []struct {
		name      string
		plainText []byte
		password  string
		wantErr   error
	}{
		{
			name:      "valid plainText and password",
			plainText: []byte("Hello, world!"),
			password:  "testpassword",
		},
		{
			name:      "empty plainText",
			plainText: []byte{},
			password:  "testpassword",
		},
		{
			name:      "invalid password (empty)",
			plainText: []byte("Hello, world!"),
			password:  "",
			wantErr:   errors.New("password is too short (min 6 chars)"),
		},
		{
			name:     "invalid password (too short)",
			password: "short",
			wantErr:  errors.New("password is too short (min 6 chars)"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cipherText, gotErr := encrypt(tt.plainText, tt.password)
			if fmt.Sprintf("%v", gotErr) != fmt.Sprintf("%v", tt.wantErr) {
				t.Fatalf("encrypt(%q, %q) err = %v expected %v", tt.plainText, tt.password, gotErr, tt.wantErr)
			}
			if gotErr != nil {
				return
			}

			gotText, err := decrypt(cipherText, tt.password)
			if err != nil {
				t.Fatalf("decrypt(%q, %q) err = %v", cipherText, tt.password, err)
			}
			if !bytes.Equal(gotText, tt.plainText) {
				t.Errorf("decrypt(%q, %q) = %q; want %q", cipherText, tt.password, gotText, tt.plainText)
			}
		})
	}
}

func TestInvalidMagicBytes(t *testing.T) {
	password := "testpassword"
	plainText := []byte("Hello, world!")

	encryptedText, err := encrypt(plainText, password)
	if err != nil {
		t.Fatalf("Error encrypting text: %v", err)
	}

	// Manipulate the magic bytes to make them invalid
	encryptedText[0] = 'X'
	encryptedText[1] = 'Y'
	encryptedText[2] = 'Z'

	_, err = decrypt(encryptedText, password)
	if err == nil {
		t.Error("Decrypt did not return an error for invalid magic bytes")
	}
}
