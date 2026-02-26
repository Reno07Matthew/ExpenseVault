package cmd

import (
	"fmt"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Create a new user account",
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

		hash, err := bcrypt.GenerateFromPassword(passBytes, bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = store.CreateUser(username, string(hash))
		if err != nil {
			return err
		}

		fmt.Println("Signup successful.")
		return nil
	},
}

func init() {
	signupCmd.Flags().StringP("username", "u", "", "Username")
}
