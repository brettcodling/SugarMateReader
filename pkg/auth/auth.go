package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/brettcodling/SugarMateReader/pkg/notify"
)

var (
	email    string
	password string
	Token    tokenResponse
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// init initialises the auth environment variables.
func init() {
	Token = tokenResponse{
		AccessToken: os.Getenv("TOKEN"),
	}
	email = os.Getenv("EMAIL")
	if email == "" {
		log.Fatal("Missing email.")
	}
	password = os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("Missing password.")
	}
}

// GetAuth gets the access token from the SugarMate oauth endpoint using user credentials.
func GetAuth() {
	if Token.RefreshToken != "" {
		refreshToken()

		if Token.AccessToken != "" {
			return
		}
	}
	jsonBody := []byte(`{"email": "` + email + `", "password": "` + password + `"}`)
	bodyReader := bytes.NewReader(jsonBody)
	resp, err := http.Post("https://sugarmate.io/oauth/web", "application/json", bodyReader)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		notify.Warning("ERROR!", "Failed Auth")
		log.Println("error:")
		log.Println("Failed Auth.")

		return
	}
	json.Unmarshal(body, &Token)
}

// refreshToken gets the access token from the SugarMate oauth endpoint using a refresh token.
func refreshToken() {
	jsonBody := []byte(`{"access_token": "` + Token.AccessToken + `", "refresh_token": "` + Token.RefreshToken + `"}`)
	bodyReader := bytes.NewReader(jsonBody)
	resp, err := http.Post("https://sugarmate.io/oauth/web/refresh", "application/json", bodyReader)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		notify.Warning("ERROR!", "Failed Auth")
		log.Println("error:")
		log.Println("Failed Auth.")

		return
	}
	json.Unmarshal(body, &Token)
}
