package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	url "net/url"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"github.com/pkg/browser"
	"github.com/ttacon/chalk"
	"golang.design/x/clipboard"
	"golang.org/x/oauth2"
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

func InitClient() *github.Client {
	godotenv.Load()

	clientId := os.Getenv("CLIENT_ID")

	clientURL := url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     "/login/device/code",
		RawQuery: fmt.Sprintf("client_id=%s&scope=%s", clientId, "repo"),
	}

	req, err := http.NewRequest(http.MethodPost, clientURL.String(), nil)
	if err != nil {
		fmt.Printf("Unable to create request: %v\n", err)
	}
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Unable to complete request: %v", err)
	}

	dcRes := ParseResponse[DeviceCodeResponse](res)

	clipboard.Init()
	clipboard.Write(clipboard.FmtText, []byte(dcRes.UserCode))
	fmt.Printf("Please enter the following code in the browser (code already copied to clipboard ehe): %s%s\n", chalk.Yellow, chalk.Bold.TextStyle(dcRes.UserCode))
	browser.OpenURL(dcRes.VerificationURI)

	pollDuration := time.Duration(dcRes.Interval) * time.Second

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Waiting for you to enter the code..."
	s.Start()

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
			if pollBody.Error == "slow_down" {
				pollDuration = pollDuration + (time.Duration(5) * time.Second)
			}

			time.Sleep(pollDuration)
			continue
		}

		if pollBody.AccessToken != "" {
			s.Stop()
			fmt.Println("✅ All cleared! Thanks for authenticating uwu!")
			os.Setenv("ACCESS_TOKEN", pollBody.AccessToken)
			break
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
