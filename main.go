package main

import (
	"log"
	"math/rand"
	"os"
	"text/template"
	"time"
)

type Rand struct {
	RandVar  string
	RandPass string
}

func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func main() {
	template, err := template.ParseFiles("payloads/ninja.payload")
	if err != nil {
		log.Fatalln(err)
	}
	data := Rand{RandVar: RandString(10), RandPass: RandString(20)}
	template.Execute(os.Stdout, data)
}
