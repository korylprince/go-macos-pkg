package macospkg_test

import (
	"crypto/rsa"
	"os"
	"testing"

	macospkg "github.com/korylprince/go-macos-pkg"
	"golang.org/x/crypto/pkcs12"
)

func TestFull(t *testing.T) {
	if os.Getenv("DEVELOPER_IDENTITY") == "" {
		t.Skip("DEVELOPER_IDENTITY not set")
	}

	identity, err := os.ReadFile(os.Getenv("DEVELOPER_IDENTITY"))
	if err != nil {
		t.Fatal("could not read identity: %w", err)
	}
	key, cert, err := pkcs12.Decode(identity, os.Getenv("DEVELOPER_IDENTITY_PASSWORD"))
	if err != nil {
		t.Fatal("could not decode identity: %w", err)
	}

	postinstall := []byte(`#!/bin/bash
echo "Hello, World!"
`)

	pkg, err := macospkg.GeneratePkg("com.github.korylprince.go-macos-pkg.test", "1.0.0", postinstall)
	if err != nil {
		t.Fatal("generate: want nil err, have: %w", err)
	}

	signedPkg, err := macospkg.SignPkg(pkg, cert, key.(*rsa.PrivateKey))
	if err != nil {
		t.Fatal("sign: want nil err, have: %w", err)
	}

	if err = macospkg.VerifyPkg(signedPkg); err != nil {
		t.Fatal("verify: want nil err, have: %w", err)
	}
}
