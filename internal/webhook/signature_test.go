package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func computeSignature(payload, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature_Valid(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"action": "published"}`)
	signature := computeSignature(payload, secret)

	if !VerifySignature(payload, signature, secret) {
		t.Error("expected valid signature to return true")
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"action": "published"}`)
	wrongSignature := "sha256=0000000000000000000000000000000000000000000000000000000000000000"

	if VerifySignature(payload, wrongSignature, secret) {
		t.Error("expected invalid signature to return false")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	secret := []byte("test-secret")
	wrongSecret := []byte("wrong-secret")
	payload := []byte(`{"action": "published"}`)
	signature := computeSignature(payload, wrongSecret)

	if VerifySignature(payload, signature, secret) {
		t.Error("expected signature with wrong secret to return false")
	}
}

func TestVerifySignature_MissingPrefix(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"action": "published"}`)
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	signatureWithoutPrefix := hex.EncodeToString(mac.Sum(nil))

	if VerifySignature(payload, signatureWithoutPrefix, secret) {
		t.Error("expected signature without sha256= prefix to return false")
	}
}

func TestVerifySignature_EmptySignature(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"action": "published"}`)

	if VerifySignature(payload, "", secret) {
		t.Error("expected empty signature to return false")
	}
}

func TestVerifySignature_InvalidHex(t *testing.T) {
	secret := []byte("test-secret")
	payload := []byte(`{"action": "published"}`)

	if VerifySignature(payload, "sha256=notvalidhex!", secret) {
		t.Error("expected invalid hex to return false")
	}
}
