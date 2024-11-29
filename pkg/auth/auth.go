package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/style"
	"github.com/brettcodling/SugarMateReader/pkg/database"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
	keyring "github.com/zalando/go-keyring"
)

var (
	email, password           string
	emailField, passwordField nucular.TextEditor
	LoginCh                   chan bool
	Token                     tokenResponse
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
	emailField.Flags = nucular.EditField
	emailField.SingleLine = true
	passwordField.Flags = nucular.EditField
	passwordField.SingleLine = true
	passwordField.PasswordChar = '*'

	LoginCh = make(chan bool)

	email = database.Get("EMAIL")
	if len(email) > 0 {
		var err error
		password, err = keyring.Get("SugarMateReader", email)
		if err != nil {
			notify.Warning("ERROR!", err.Error())
			OpenLogin()
		} else if password == "" {
			OpenLogin()
		}
	} else {
		OpenLogin()
	}
}

// OpenLogin will open the login window
func OpenLogin() {
	emailField.SelectAll()
	emailField.Text([]rune(email))
	passwordField.SelectAll()
	passwordField.Text([]rune(password))
	wnd := nucular.NewMasterWindow(0, "Login", updateLogin)
	wnd.SetStyle(style.FromTheme(style.DarkTheme, 2.0))
	wnd.Main()
}

// updateLogin will setup the login window and wait for updates
// when the login button is clicked the email will be stored in a file and the password in keyring
func updateLogin(w *nucular.Window) {
	w.Row(50).Dynamic(1)
	w.Label("Email:", "LC")
	emailField.Edit(w)
	w.Row(50).Dynamic(1)
	w.Label("Password:", "LC")
	passwordField.Edit(w)
	w.Row(50).Dynamic(1)
	if w.ButtonText("Login") {
		email = string(emailField.Buffer)
		password = string(passwordField.Buffer)
		err := keyring.Set("SugarMateReader", email, password)
		if err != nil {
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("EMAIL", email)
		w.Master().Close()
		go func() {
			LoginCh <- true
		}()
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
	resp, err := http.Post("https://api.sugarmate.io/oauth/web", "application/json", bodyReader)
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
	resp, err := http.Post("https://api.sugarmate.io/oauth/web/refresh", "application/json", bodyReader)
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
