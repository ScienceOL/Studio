// nolint:revive
package utils

import (
	"errors"
	"reflect"

	"github.com/golang-jwt/jwt/v5"
	"github.com/scienceol/studio/service/pkg/common/code"
)

const (
	// DefaultPrivateKey RSA 私钥 (PEM 格式)
	DefaultPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIBIjANBgkqhkiG9...你的私钥内容...
-----END RSA PRIVATE KEY-----`

	// DefaultPublicKey RSA 公钥 (PEM 格式)
	DefaultPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnn3jPyW81YqSjSLWBkdE
ZzurZ5gimj6Db693bO0WvhMPABpYdOTeAU1mnQh2ep4H7zoUdz4PKARh/p5Meh6l
ejtbyliptvW9WXg5LoquIzPyTe5/2W9GoTrzDHMdM89Gc2dn16TbsKU5z3lROlBP
Q2v7UjQCbs8VpSogb44kOn0cx/MV2+VBfJzFWkJnaXxc101YUteJytJRMli0Wqev
nYqzCgrtbdvqVF/8hqETZOIWdWlhRDASdYw3R08rChcMJ9ucZL/VUM+aKu+feekQ
UZ6Bi6CeZjgqBoiwccApVR88WbyVXWR/3IFvJb0ndoSdH85klpp25yVAHTdSIDZP
lQIDAQAB
-----END PUBLIC KEY-----`
)

type PayLoad struct {
	UserID uint64 `json:"userId"`
	Email  string `json:"email"`
	OrgID  uint64 `json:"orgId"`
}

type Claims struct {
	Identity PayLoad `json:"identity"`
	jwt.RegisteredClaims
}

// ParseJWT 解析 JWT token (使用默认公钥)
func ParseJWT(tokenString string, claims jwt.Claims) error {
	return ParseJWTWithPublicKey(tokenString, DefaultPublicKey, claims)
}

func ParseJWTWithPublicKey(tokenString, publicKeyPEM string, claims jwt.Claims) error {
	// 解析公钥
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		return err
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// 验证签名方法是否为 RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return err
	}

	if reflect.TypeOf(token.Claims) != reflect.TypeOf(claims) {
		return code.InvalidateJWT
	}

	return nil
}

func GenerateJWTWithPrivateKey(claims jwt.Claims, privateKeyPEM string) (string, error) {
	// 解析私钥
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// ParseJWTWithSecret 保持兼容性 - 使用 HS256
func ParseJWTWithSecret(tokenString, secret string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return err
	}

	if reflect.TypeOf(token.Claims) != reflect.TypeOf(claims) {
		return code.InvalidateJWT
	}

	return nil
}

// GenerateJWTWithSecret 保持兼容性 - 使用 HS256
func GenerateJWTWithSecret(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
