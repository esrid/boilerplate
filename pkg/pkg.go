package pkg

import (
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func CleanAndLower(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

func GenerateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func GetIPAddressBytes(r *http.Request) []byte {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		ip := strings.TrimSpace(ips[0])
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil {
			return parsedIP
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil {
			return parsedIP
		}
	}
	return nil
}
