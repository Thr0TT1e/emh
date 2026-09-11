// Утилита для генерации bcrypt-хеша пароля администратора.
// Использование: go run ./cmd/hashpassword <password>
package main

import (
	"fmt"
	"os"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: hashpassword <password>")
		os.Exit(1)
	}
	hash, err := auth.HashPassword(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(hash)
}
