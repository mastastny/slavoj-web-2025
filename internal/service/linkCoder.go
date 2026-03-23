package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/mastastny/slavoj-web-2025/internal/config"
)

type LinkCoder struct {
	conf config.Config
}

func NewLinkCoder(conf config.Config) *LinkCoder {
	return &LinkCoder{conf: conf}
}

func (lc *LinkCoder) Encode(id int64) string {
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(id))

	mac := hmac.New(sha256.New, []byte(lc.conf.LinkSecretKey))
	mac.Write(msg)

	payload := append(msg, mac.Sum(nil)...)
	return base64.URLEncoding.EncodeToString(payload)
}

func (lc *LinkCoder) Decode(encoded string) (int64, error) {
	payload, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, fmt.Errorf("linkCoder.Decode: %w", err)
	}
	if len(payload) != 8+32 {
		return 0, errors.New("linkCoder.Decode: invalid payload length")
	}

	msg := payload[:8]
	sig := payload[8:]

	mac := hmac.New(sha256.New, []byte(lc.conf.LinkSecretKey))
	mac.Write(msg)

	if !hmac.Equal(sig, mac.Sum(nil)) {
		return 0, errors.New("linkCoder.Decode: invalid signature")
	}

	return int64(binary.BigEndian.Uint64(msg)), nil
}
