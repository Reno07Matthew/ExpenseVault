package cmd

import (
	"fmt"
	"os"
	"path/filepath"


	"github.com/golang-jwt/jwt/v5"
)

// getCurrentUserID retrieves the logged-in user's ID.
func getCurrentUserID() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	path := filepath.Join(home, ".expensevault", "token")
	tokenBytes, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("not logged in: %w", err)
	}

	tokenString := string(tokenBytes)
	jwtSecret := []byte("your-secret-key-change-in-production")
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["sub"].(string)
		user, err := store.GetUserByUsername(username)
		if err != nil {
			return 0, fmt.Errorf("user not found: %w", err)
		}
		return user.ID, nil
	}

	return 0, fmt.Errorf("invalid token claims")
}
