package jwt

import (
	"testing"
)

func TestJWT(t *testing.T) {
	jwt, exp, iat := Create(1, "CREATOR")

	validClaims := Claims{
		Sub: 1,
		Rol: "CREATOR",
		Exp: exp,
		Iat: iat,
	}

	if exp != iat+TokenLifetime {
		t.Fatalf("INVALID EXPIRATION TIME. EXPECTED %d GOT %d", iat+TokenLifetime, exp)
	}

	_, claims, err := Parse(jwt)
	if err != nil {
		t.Fatalf("UNEXPECTED ERROR: %v", err)
	}

	if claims != validClaims {
		t.Fatalf("INVALID CLAIMS. EXPECTED %v GOT %v", validClaims, claims)
	}
}
