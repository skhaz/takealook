package session

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/argon2"
)

var secret = []byte(os.Getenv("JWT_SECRET"))

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return []byte{}, err
	}

	return salt, nil
}

func HashPassword(password string, salt []byte) (string, error) {
	time := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLength := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLength)

	return base64.StdEncoding.EncodeToString(salt) + "$" + base64.StdEncoding.EncodeToString(hash), nil
}

func ComparePasswords(storedPassword, inputPassword string) (bool, error) {
	parts := strings.Split(storedPassword, "$")
	if len(parts) != 2 {
		return false, fmt.Errorf("stored password is not in the correct format")
	}

	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	inputHash, err := HashPassword(inputPassword, salt)
	if err != nil {
		return false, err
	}

	inputHashParts := strings.Split(inputHash, "$")
	if len(inputHashParts) != 2 {
		return false, fmt.Errorf("input hash generation failed")
	}

	storedHash := parts[1]
	inputGeneratedHash := inputHashParts[1]

	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(inputGeneratedHash)) == 1 {
		return true, nil
	}

	return false, nil
}

func SetCookie(email string, c echo.Context) error {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: expiration.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(secret)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = signedString
	cookie.Expires = expiration
	cookie.HttpOnly = true
	c.SetCookie(cookie)

	return nil
}

func VerifyCookie(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			return c.Redirect(http.StatusFound, "/join")
		}

		tokenString := cookie.Value
		claims := &jwt.StandardClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			return c.Redirect(http.StatusFound, "/join")
		}

		c.Set("email", claims.Subject)

		return next(c)
	}
}

func SkipPassword(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			return next(c)
		}

		tokenString := cookie.Value
		claims := &jwt.StandardClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			return next(c)
		}

		c.Set("email", claims.Subject)

		return c.Redirect(http.StatusFound, "/dashboard")
	}
}

func Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.MaxAge = -1
	c.SetCookie(cookie)

	return c.Redirect(http.StatusFound, "/")
}
