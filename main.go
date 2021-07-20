package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const IpifyApi string = "https://api.ipify.org"

func handleResponse(resp *http.Response, err error, errorCode int) {
	if err != nil {
		log.Println(err.Error())
		os.Exit(errorCode)
	}

	if resp.StatusCode > 200 {
		bodyText, _ := ioutil.ReadAll(resp.Body)
		log.Printf("%v - %v", resp.StatusCode, string(bodyText))
		os.Exit(errorCode)
	}
}

func main() {
	domain := flag.String("d", "", "Domain to update")
	flag.Parse()

	if len(*domain) == 0 {
		log.Println("Domain is required")
		os.Exit(ERR_ARGS)
	}

	// Get IP
	resp, err := http.Get(IpifyApi)
	handleResponse(resp, err, ERR_GET_IP)

	defer resp.Body.Close()

	ipBody, ipErr := ioutil.ReadAll(resp.Body)
	if ipErr != nil {
		log.Println(ipErr.Error())
		os.Exit(ERR_GET_IP)
	}
	ip := string(ipBody)

	// Hover login
	user := os.Getenv("HOVER_USER")
	password := os.Getenv("HOVER_PASSWORD")
	loginJson := fmt.Sprintf("{\"username\":\"%s\", \"password\":\"%s\"}", user, password)

	resp, err = http.Post(HoverLoginApi, "application/json", bytes.NewBufferString(loginJson))
	handleResponse(resp, err, ERR_HOVER_LOGIN)

	// Create an auth client
	client, err := CreateHoverClient(resp.Cookies())
	if err != nil {
		log.Println(err.Error())
		os.Exit(ERR_HOVER_CLIENT)
	}

	// Get dns records
	req, err := http.NewRequest("GET", fmt.Sprintf(HoverGetDnsApi, *domain), nil)
	if err != nil {
		log.Println(err.Error())
		os.Exit(ERR_HOVER_GET_DNS)
	}

	resp, err = client.Do(req)
	handleResponse(resp, err, ERR_HOVER_GET_DNS)

	dnsBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		os.Exit(ERR_HOVER_GET_DNS)
	}

	dnsResult := HoverDnsResult{}

	err = json.Unmarshal(dnsBody, &dnsResult)
	if err != nil {
		log.Println(err.Error())
		os.Exit(ERR_HOVER_GET_DNS)
	}

	if !dnsResult.Succeeded {
		log.Printf("%v - %v - %v", resp.StatusCode, dnsResult.ErrorCode, dnsResult.Error)
		os.Exit(ERR_HOVER_GET_DNS)
	}

	var atARecordId string
	for _, entry := range dnsResult.Domains[0].Entries {
		if entry.Type == "A" && entry.Name == "@" {
			atARecordId = entry.Id
			break
		}
	}

	if len(atARecordId) == 0 {
		log.Println("Cannot find the @ A Record")
		os.Exit(ERR_HOVER_GET_DNS)
	}

	// Update the DNS
	// ip = "127.0.0.1"
	// updateJson := fmt.Sprintf("{\"content\":\"%s\"}", ip)
	payload := HoverDnsUpdatePayload{
		Domain: HoverDnsUpdateDomain{
			Id: "domain-" + *domain,
			Records: []HoverDnsUpdateRecords{
				{
					Id: atARecordId,
				},
			},
		},
		Fields: HoverDnsUpdateFields{
			Content: ip,
		},
	}
	updateJson, _ := json.Marshal(payload)
	req, err = http.NewRequest("PUT", HoverUpdateDnsApi, bytes.NewBuffer(updateJson))

	if err != nil {
		log.Println(err.Error())
		os.Exit(ERR_HOVER_UPDATE_DNS)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	handleResponse(resp, err, ERR_HOVER_UPDATE_DNS)

	log.Printf("%v - DNS A record updated with the IP: %s", resp.StatusCode, ip)
}
