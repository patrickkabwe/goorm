package goorm

import (
	"fmt"
	"os"
)

const ()

func GetENV(key string) string {
	fmt.Println(os.Getenv(key))
	return os.Getenv(key)
}
