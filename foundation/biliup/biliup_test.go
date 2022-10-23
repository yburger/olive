package biliup_test

import (
	"fmt"

	"github.com/go-olive/olive/foundation/biliup"
)

func ExampleBiliup_Upload() {
	err := biliup.New(biliup.Config{
		CookieFilepath:    "",
		VideoFilepath:     "",
		Threads:           6,
		MaxBytesPerSecond: 2097152,
	}).Upload()
	if err != nil {
		fmt.Println(err)
	}
}
