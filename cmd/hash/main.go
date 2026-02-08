package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pass := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	fmt.Println(string(hash))
}
