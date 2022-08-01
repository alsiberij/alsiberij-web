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

	RoleCreator        = "CREATOR"
	RoleAdmin          = "ADMIN"
	RoleModerator      = "MODERATOR"
	RolePrivilegedUser = "PRIVILEGED_USER"
	RoleUser           = "USER"
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

	CanBeBannedByModerator = []string{RoleUser}
	CanBeBannedByAdmin     = []string{RoleModerator, RolePrivilegedUser, RoleUser}
	CanBeBannedByCreator   = []string{RoleAdmin, RoleModerator, RolePrivilegedUser, RoleUser}
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
		Exp: issueTime + JwtLifetime,
		Iat: issueTime,
	}

	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)

	base64enc := base64.URLEncoding.WithPadding(base64.NoPadding)

	headerString := base64enc.EncodeToString(headerBytes)
	claimsString := base64enc.EncodeToString(claimsBytes)

	token := headerString + "." + claimsString

	h := hmac.New(sha256.New, key)
	h.Write([]byte(token))
	signature := base64enc.EncodeToString(h.Sum(nil))

	return headerString + "." + claimsString + "." + signature,
		issueTime + JwtLifetime,
		issueTime
}

func Parse(jwt string) (Header, Claims, error) {
	jwtParts := strings.Split(jwt, ".")
	if len(jwtParts) != 3 {
		return Header{}, Claims{}, errorInvalidJwtParts
	}

	base64enc := base64.URLEncoding.WithPadding(base64.NoPadding)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(jwtParts[0] + "." + jwtParts[1]))
	signature := base64enc.EncodeToString(h.Sum(nil))

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
