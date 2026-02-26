package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"expenseVault/api"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to an existing account",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		if username == "" {
			return fmt.Errorf("username is required")
		}

		fmt.Print("Password: ")
		passBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			return err
		}

		user, err := store.GetUserByUsername(username)
		if err != nil {
			return err
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), passBytes); err != nil {
			return fmt.Errorf("invalid username or password")
		}

		_ = store.UpdateLastLogin(user.ID)

		jwtSecret := []byte("your-secret-key-change-in-production")
		token, err := api.GenerateToken(username, jwtSecret)
		if err != nil {
			return err
		}

		path, err := saveToken(token)
		if err != nil {
			return err
		}

		fmt.Printf("Login successful. Token saved to %s\n", path)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username")
}

func saveToken(token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".expensevault")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "token")
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return "", err
	}
	return path, nil
}
