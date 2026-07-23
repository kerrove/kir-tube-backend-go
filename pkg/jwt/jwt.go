package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Secret string
}

type JWTData struct {
	Id      string
	IsAdmin bool
}

func NewJWT(secret string) *JWT {
	return &JWT{Secret: secret}
}

func (j *JWT) Create(data JWTData, expiresIn string) (string, error) {
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return "", err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      data.Id,
		"isAdmin": data.IsAdmin,
		"exp":     time.Now().Add(duration).Unix(),
	})

	s, err := t.SignedString([]byte(j.Secret))
	if err != nil {
		return "", err
	}
	return s, nil
}

func (j *JWT) Parse(token string) (bool, *JWTData) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(j.Secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return false, nil
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return false, nil
	}

	data := &JWTData{}

	if id, ok := claims["id"].(string); ok {
		data.Id = id
	}

	if isAdmin, ok := claims["isAdmin"].(bool); ok {
		data.IsAdmin = isAdmin
	}

	return true, data
}
