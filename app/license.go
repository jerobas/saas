package main

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

type LicensePayload struct {
	U string `json:"userId"`
	E string `json:"email"`
	I string `json:"issuedAt"`
	X string `json:"expiresAt"`
}

func ValidateLicenseWithKey(license string, publicKey interface{}) (*LicensePayload, error) {
	parts := strings.Split(license, ".")
	if len(parts) != 2 {
		return nil, errors.New("license format invalid")
	}

	payload64 := parts[0]
	sig64 := parts[1]

	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(publicKey.(ed25519.PublicKey), []byte(payload64), sig) {
		return nil, errors.New("invalid signature")
	}

	packed, err := base64.StdEncoding.DecodeString(payload64)
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		return nil, err
	}

	raw, _ := io.ReadAll(reader)

	var payload LicensePayload
	json.Unmarshal(raw, &payload)

	expTime, err := time.Parse(time.RFC3339, payload.X)
	if err != nil {
		return nil, err
	}

	if time.Now().After(expTime) {
		return nil, errors.New("license expired")
	}

	return &payload, nil
}
