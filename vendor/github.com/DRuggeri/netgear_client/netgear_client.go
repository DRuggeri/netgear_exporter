package netgear_client

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type NetgearClient struct {
	http_client *http.Client
	router_url  string
	cookie      string
	sessionid   string
	insecure    bool
	username    string
	password    string
	timeout     int
	debug       bool
}

type SOAPEnvelope struct {
	XMLName  xml.Name
	SOAPBody struct {
		XMLName         xml.Name
		ResponseCode    string `xml:"ResponseCode"`
		ResponseContent []byte `xml:",innerxml"`
	} `xml:"Body"`
}

type Node struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
	Nodes   []Node `xml:",any"`
}

const SOAP_PATH = "/soap/server_sa/"
const SESSION_ID = "A7D88AE69687E58D9A00"
const AUTH_ACTION = "urn:NETGEAR-ROUTER:service:DeviceConfig:1#SOAPLogin"
const AUTH_REQUEST = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<SOAP-ENV:Envelope
  xmlns:SOAPSDK1="http://www.w3.org/2001/XMLSchema"
  xmlns:SOAPSDK2="http://www.w3.org/2001/XMLSchema-instance"
  xmlns:SOAPSDK3="http://schemas.xmlsoap.org/soap/encoding/"
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
  <SOAP-ENV:Header>
    <SessionID>%s</SessionID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <M1:SOAPLogin xmlns:M1="urn:NETGEAR-ROUTER:service:DeviceConfig:1">
      <Username>%s</Username>
      <Password>%s</Password>
    </M1:SOAPLogin>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

func NewNetgearClient(i_url string, i_insecure bool, i_username string, i_password string, i_timeout int, i_debug bool) (*NetgearClient, error) {
	if i_debug {
		log.Printf("netgear_client.go: Constructing debug client\n")
	}

	if i_url == "" {
		i_url = "https://routerlogin.net"
	}
	if i_username == "" {
		i_username = "admin"
	}
	if i_password == "" {
		return nil, errors.New("Admin password is required")
	}

	if strings.HasSuffix(i_url, "/") {
		i_url = i_url[:len(i_url)-1]
	}
	if !strings.Contains(i_url, "://") {
		i_url = "https://" + i_url
	}

	_, err := url.Parse(i_url)
	if err != nil {
		return nil, fmt.Errorf("Error parsing provided URL (%s): %v", i_url, err)
	}

	/* Disable TLS verification if requested */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: i_insecure},
	}

	client := NetgearClient{
		http_client: &http.Client{
			Timeout:   time.Second * time.Duration(i_timeout),
			Transport: tr,
		},
		router_url: i_url,
		insecure:   i_insecure,
		username:   i_username,
		password:   i_password,
		cookie:     "UNSET",
		sessionid:  SESSION_ID,
		debug:      i_debug,
	}

	return &client, nil
}

func (client *NetgearClient) LogIn() error {
	_, err := client.send_request(AUTH_ACTION, fmt.Sprintf(AUTH_REQUEST, client.sessionid, client.username, client.password), false)
	return err
}

func (client *NetgearClient) send_request(action string, data string, attempt_login bool) ([]byte, error) {
	tries := 2
	if !attempt_login {
		tries = 1
	}

	full_url := client.router_url + SOAP_PATH

	for i := 0; i < tries; i++ {
		if client.debug {
			log.Printf("netgear_client.go: full url (derived)='%s', data='%s'\n", full_url, data)
		}

		req, err := http.NewRequest("POST", full_url, bytes.NewBuffer([]byte(data)))
		if err != nil {
			log.Fatal(err)
			return []byte{}, err
		}

		req.Header.Set("Content-Type", "text/xml;charset=utf-8")
		req.Header.Set("SOAPAction", action)
		req.Header.Set("Host", "routerlogin.net")
		req.Header.Set("Cookie", client.cookie)
		req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
		req.Header.Set("User-Agent", "curl/7.59.0")

		if client.debug {
			log.Printf("netgear_client.go: Sending HTTP request to %s...\n", req.URL)
			log.Printf("netgear_client.go: Request headers:\n")
			for name, headers := range req.Header {
				for _, h := range headers {
					log.Printf("netgear_client.go:   %v: %v", name, h)
				}
			}

			log.Printf("netgear_client.go: BODY:\n")
			body := "<none>"
			if req.Body != nil {
				body = string(data)
			}
			log.Printf("%s\n", body)
		}

		resp, err := client.http_client.Do(req)

		if err != nil {
			return []byte{}, err
		}

		client.cookie = resp.Header.Get("Set-Cookie")

		if client.debug {
			log.Printf("netgear_client.go: Response code: %d\n", resp.StatusCode)
			log.Printf("netgear_client.go: Response headers:\n")
			for name, headers := range resp.Header {
				for _, h := range headers {
					log.Printf("netgear_client.go:   %v: %v", name, h)
				}
			}
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return []byte{}, err
		}

		var response SOAPEnvelope
		err = xml.Unmarshal(bodyBytes, &response)
		if err != nil {
			return []byte{}, err
		}

		if client.debug {
			log.Printf("netgear_client.go: BODY:\n%s\n\n\n", string(bodyBytes))
		}

		if response.SOAPBody.ResponseCode == "401" {
			/* We're on our first try and we found a not logged in situation */
			if attempt_login && i == 0 {
				if client.debug {
					log.Printf("netgear_client.go: Detected client not being logged in. Executing login...\n\n\n")
				}
				err = client.LogIn()
				if err != nil {
					return []byte{}, err
				}
				continue
			} else {
				return []byte{}, errors.New("The netgear_client is not logged in!")
			}
		}

		return response.SOAPBody.ResponseContent, nil
	}

	return []byte{}, fmt.Errorf("Failed to execute request after %d tries", tries)
}
