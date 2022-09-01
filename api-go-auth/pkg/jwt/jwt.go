package jwt

import (
	"auth/pkg/utils"
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
	TokenLifetime = 3600
)

type (
	Header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	Claims struct {
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

	claims := Claims{
		Sub: userId,
		Rol: role,
		Exp: issueTime + TokenLifetime,
		Iat: issueTime,
	}

	hBytes, _ := json.Marshal(header)
	cBytes, _ := json.Marshal(claims)

	enc := base64.URLEncoding.WithPadding(base64.NoPadding)
	hsh := sha256.New()

	hEncSize, cEncSize, sEncSize := enc.EncodedLen(len(hBytes)), enc.EncodedLen(len(cBytes)), enc.EncodedLen(hsh.Size())
	jwtSize := hEncSize + 1 + cEncSize + 1 + sEncSize

	buf := make([]byte, jwtSize)

	buf[hEncSize] = '.'
	buf[hEncSize+1+cEncSize] = '.'

	enc.Encode(buf[:hEncSize], hBytes)
	enc.Encode(buf[hEncSize+1:], cBytes)

	h := hmac.New(sha256.New, key)
	h.Write(buf[:hEncSize+1+cEncSize])

	enc.Encode(buf[hEncSize+1+cEncSize+1:], h.Sum(nil))

	return utils.BytesToString(buf),
		issueTime + TokenLifetime,
		issueTime
}

func Parse(jwt string) (Header, Claims, error) {
	jwtParts := strings.Split(jwt, ".")
	if len(jwtParts) != 3 {
		return Header{}, Claims{}, errorInvalidJwtParts
	}

	enc := base64.URLEncoding.WithPadding(base64.NoPadding)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(jwtParts[0] + "." + jwtParts[1]))
	signature := enc.EncodeToString(h.Sum(nil))

	if signature != jwtParts[2] {
		return Header{}, Claims{}, errorInvalidSignature
	}

	if i := len(jwtParts[0]) % 4; i != 0 {
		jwtParts[0] += strings.Repeat("=", 4-i)
	}

	headerBytes, err := base64.URLEncoding.DecodeString(jwtParts[0])
	if err != nil {
		return Header{}, Claims{}, err
	}

	var header Header
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return Header{}, Claims{}, err
	}

	if i := len(jwtParts[1]) % 4; i != 0 {
		jwtParts[1] += strings.Repeat("=", 4-i)
	}

	claimsBytes, err := base64.URLEncoding.DecodeString(jwtParts[1])
	if err != nil {
		return Header{}, Claims{}, err
	}

	var claims Claims
	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		return Header{}, Claims{}, err
	}

	if time.Unix(claims.Exp, 0).Before(time.Now()) {
		return Header{}, Claims{}, errorExpired
	}

	return header, claims, nil
}
