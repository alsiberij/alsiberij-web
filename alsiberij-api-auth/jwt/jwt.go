package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	JwtLifetime = 3600
)

type (
	Header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	Payload struct {
		Sub int64  `json:"sub"`
		Rol string `json:"rol"`
		Exp int64  `json:"exp"`
		Iat int64  `json:"iat"`
	}
)

var (
	//go:embed SECRET.KEY
	key []byte

	errorInvalidJwtParts  = errors.New("JWT should contain 3 parts")
	errorInvalidSignature = errors.New("JWT signature invalid")
	errorExpired          = errors.New("JWT expired")
)

func Create(userId int64, role string) (string, int64, int64) {
	header := Header{
		Alg: "HS256",
		Typ: "JWT",
	}

	issueTime := time.Now().Unix()

	payload := Payload{
		Sub: userId,
		Rol: role,
		Exp: issueTime + JwtLifetime,
		Iat: issueTime,
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	base64enc := base64.URLEncoding.WithPadding(base64.NoPadding)

	headerString := base64enc.EncodeToString(headerBytes)
	payloadString := base64enc.EncodeToString(payloadBytes)

	token := headerString + "." + payloadString

	h := hmac.New(sha256.New, key)
	h.Write([]byte(token))
	signature := base64enc.EncodeToString(h.Sum(nil))

	return headerString + "." + payloadString + "." + signature,
		issueTime + JwtLifetime,
		issueTime
}

func Parse(jwt string) (Header, Payload, error) {
	jwtParts := strings.Split(jwt, ".")
	if len(jwtParts) != 3 {
		return Header{}, Payload{}, errorInvalidJwtParts
	}

	base64enc := base64.URLEncoding.WithPadding(base64.NoPadding)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(jwtParts[0] + "." + jwtParts[1]))
	signature := base64enc.EncodeToString(h.Sum(nil))

	if signature != jwtParts[2] {
		return Header{}, Payload{}, errorInvalidSignature
	}

	if i := len(jwtParts[0]) % 4; i != 0 {
		jwtParts[0] += strings.Repeat("=", 4-i)
	}
	headerBytes, err := base64.URLEncoding.DecodeString(jwtParts[0])
	if err != nil {
		return Header{}, Payload{}, err
	}
	var header Header
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return Header{}, Payload{}, err
	}

	if i := len(jwtParts[1]) % 4; i != 0 {
		jwtParts[1] += strings.Repeat("=", 4-i)
	}
	payloadBytes, err := base64.URLEncoding.DecodeString(jwtParts[1])
	if err != nil {
		return Header{}, Payload{}, err
	}
	var payload Payload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return Header{}, Payload{}, err
	}

	if time.Unix(payload.Exp, 0).Before(time.Now()) {
		return Header{}, Payload{}, errorExpired
	}

	return header, payload, nil
}
