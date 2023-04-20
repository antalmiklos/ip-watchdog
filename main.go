package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

type IPResponse struct {
	IP string `json:"ip"`
}

var ip string

var appPass string
var gmailUser string
var frequency int

const (
	smtpServer = "smtp.gmail.com"
	smtpPort   = "465"
)

func init() {
	var err error
	pf := os.Getenv("IP_WATCHDOG_POLL_FREQUENCY")
	frequency, err = strconv.Atoi(pf)
	if err != nil {
		log.Fatal("invalid frequency: ", pf)
	}

	appPass = os.Getenv("IP_WATCHDOG_GMAIL_PASS")
	gmailUser = os.Getenv("IP_WATCHDOG_GMAIL_USER")
}

func getPublicIP() (string, error) {
	response, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var ipData IPResponse
	err = json.Unmarshal(body, &ipData)
	if err != nil {
		return "", err
	}

	return ipData.IP, nil
}

func compIps(new string) bool {
	return ip == new
}

func sendMail() error {
	subject := "Subject: !! IP CHANGE !!\n"
	body := "New IP: " + ip

	msg := []byte(subject + "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n" + body)

	auth := smtp.PlainAuth("", gmailUser, appPass, smtpServer)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpServer,
	}

	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(gmailUser); err != nil {
		return err
	}

	if err = client.Rcpt(gmailUser); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(msg); err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = client.Quit()
	if err != nil {
		return err
	}

	return nil
}
func worker() {
	publicIP, err := getPublicIP()
	if err != nil {
		fmt.Printf("Error getting public IP: %s\n", err)
		return
	}

	if !compIps(publicIP) {
		ip = publicIP
		sendMail()
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("My public IP address is: %s\n", ip)
}

func main() {
	worker()

	ticker := time.NewTicker(time.Duration(frequency) * time.Minute)
	for {
		select {
		case <-ticker.C:
			worker()
			time.Sleep(100 * time.Millisecond)
		}
	}
}
