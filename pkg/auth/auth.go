package auth

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	sessionCookieName = "albumpedia_session"
	pbkdf2Iterations  = 210000
	pbkdf2KeyLen      = 32
	saltLen           = 16
)

func getSessionSecret() []byte {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "albumpedia-dev-secret-change-in-env"
	}
	return []byte(secret)
}

// HashPassword genera un hash seguro PBKDF2-HMAC-SHA256 con un salt aleatorio de 16 bytes.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, pbkdf2KeyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s$%s", hex.EncodeToString(salt), hex.EncodeToString(dk)), nil
}

// VerifyPassword compara en tiempo constante una contraseña contra su hash almacenado.
func VerifyPassword(password, storedHash string) bool {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expectedKey, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	dk, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, pbkdf2KeyLen)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(dk, expectedKey) == 1
}

func signUserID(userID int32) string {
	payload := strconv.Itoa(int(userID))
	mac := hmac.New(sha256.New, getSessionSecret())
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", payload, sig)
}

func verifySignedUserID(token string) (int32, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, false
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return 0, false
	}
	expectedToken := signUserID(int32(id))
	if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
		return 0, false
	}
	return int32(id), true
}

// SetSessionCookie emite una cookie HttpOnly firmada con HMAC-SHA256.
func SetSessionCookie(w http.ResponseWriter, userID int32) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    signUserID(userID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 días
	})
}

// ClearSessionCookie invalida la cookie de sesión en el navegador.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// GetSessionUserID obtiene y valida el ID del usuario autenticado desde la cookie.
func GetSessionUserID(r *http.Request) (int32, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return 0, false
	}
	return verifySignedUserID(cookie.Value)
}
