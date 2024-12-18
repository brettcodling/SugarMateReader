package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	_ "embed"

	"github.com/brettcodling/SugarMateReader/internal/database"
	"github.com/brettcodling/SugarMateReader/internal/notify"
	keyring "github.com/zalando/go-keyring"
)

var (
	Email, Password string
	Token           TokenResponse
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// init initialises the auth environment variables.
func init() {
	Token = TokenResponse{
		AccessToken: os.Getenv("TOKEN"),
	}

	Email = database.Get("EMAIL")
}

func LoadPassword() error {
	var err error
	Password, err = keyring.Get("SugarMateReader", Email)

	return err
}

// GetAuth gets the access token from the SugarMate oauth endpoint using user credentials.
func GetAuth() {
	if Token.RefreshToken != "" {
		refreshToken()

		if Token.AccessToken != "" {
			return
		}
	}
	jsonBody := []byte(`{"email": "` + Email + `", "password": "` + Password + `"}`)
	bodyReader := bytes.NewReader(jsonBody)
	resp, err := http.Post("https://api.sugarmate.io/oauth/web", "application/json", bodyReader)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	parseTokenBody(resp)
}

func parseTokenBody(resp *http.Response) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		notify.Warning("ERROR!", "Failed Auth")
		log.Println("error:")
		log.Println(err)

		return
	}
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
	resp, err := http.Post("https://api.sugarmate.io/oauth/web/refresh", "application/json", bodyReader)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	parseTokenBody(resp)
}
