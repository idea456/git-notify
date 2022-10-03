package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/browser"
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type DevicePollResponse struct {
	Error       string `json:"error"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type AuthResponse interface {
	DeviceCodeResponse | DevicePollResponse
}

func ParseResponse[T AuthResponse](r *http.Response) T {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Unable to read response body: %v\n", err)
	}

	var body T
	json.Unmarshal(bodyBytes, &body)

	return body
}

func main() {
	clientId := os.Getenv("CLIENT_ID")
	// scope := os.Getenv("SCOPE")

	url := url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     "/login/device/code",
		RawQuery: fmt.Sprintf("client_id=%s&scope=%s", clientId, "repo"),
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), nil)
	if err != nil {
		fmt.Printf("Unable to create request: %v\n", err)
	}
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Unable to complete request: %v", err)
	}

	dcRes := ParseResponse[DeviceCodeResponse](res)
	fmt.Printf("Please enter the following code in the browser: %s\n", dcRes.UserCode)
	browser.OpenURL(dcRes.VerificationURI)

	pollDuration := time.Duration(dcRes.Interval) * time.Second

	for {
		pollURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code", clientId, dcRes.DeviceCode)
		pollReq, err := http.NewRequest(http.MethodPost, pollURL, nil)
		pollReq.Header.Set("Accept", "application/json")
		if err != nil {
			fmt.Printf("Unable to create poll request: %v\n", err)
		}

		pollRes, err := http.DefaultClient.Do(pollReq)
		if err != nil {
			fmt.Printf("Unable to poll: %v\n", err)
		}
		pollBody := ParseResponse[DevicePollResponse](pollRes)
		if pollBody.Error != "" {
			fmt.Println(pollBody.Error)
			if pollBody.Error == "slow_down" {
				pollDuration = pollDuration + (time.Duration(5) * time.Second)
			}

			fmt.Println("waiting...")
			time.Sleep(pollDuration)
			continue
		}

		if pollBody.AccessToken != "" {
			fmt.Printf("Received token: %v\n", pollBody.AccessToken)
			os.Setenv("ACCESS_TOKEN", pollBody.AccessToken)
			break
		}
	}
}
