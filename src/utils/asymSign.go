package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"

	"reflect"
)

func AsymSign(hash []byte) (string, error) {
	path := "/private/private.pem"
	privateKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("private.pem file not found: " + path)
	}

	// Decode the key into a "block"
	privateBlock, _ := pem.Decode(privateKeyBytes)
	if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	// Parse the private key from the block
	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return "", errors.New("failed to parse private key type: %s" + err.Error())
	}

	// Check the type of the key
	if _, ok := privateKey.(*rsa.PrivateKey); !ok {
		return "", errors.New("invalid key type: " + reflect.TypeOf(privateKey).String())
	}

	// Sign the hash with the client's private key using PSS
	signatureRaw, err := rsa.SignPSS(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hash[:], &rsa.PSSOptions{
		SaltLength: 32,
		Hash:       crypto.SHA256,
	})
	if err != nil {
		return "", errors.New("signing error: %s" + err.Error())
	}

	// Encode the signature to base64 for easy transport
	return base64.StdEncoding.EncodeToString(signatureRaw), err
}
