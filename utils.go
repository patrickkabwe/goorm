package goorm

import (
	"os"
)

const ()

func GetENV(key string) string {
	return os.Getenv(key)
}
