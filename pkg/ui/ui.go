package ui

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"

	_ "embed"

	"github.com/brettcodling/SugarMateReader/pkg/auth"
	"github.com/brettcodling/SugarMateReader/pkg/database"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
	"github.com/pkg/browser"
	keyring "github.com/zalando/go-keyring"
)

var (
	//go:embed close.tmpl
	closeTmpl string
	//go:embed layout.tmpl
	layoutTmpl string
	//go:embed login.tmpl
	loginTmpl string
	//go:embed settings.tmpl
	settingsTmpl string
	url          string
)

// init initialises the auth environment variables.
func init() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Fatal(err)
	}
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/settings", handleSettings)
	go http.Serve(listener, nil)
	url = fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)

	if auth.Email != "" {
		err := auth.LoadPassword()
		if err != nil {
			notify.Warning("ERROR!", err.Error())
			OpenLogin()
			<-auth.LoginCh
		} else if auth.Password == "" {
			OpenLogin()
			<-auth.LoginCh
		}
	} else {
		OpenLogin()
		<-auth.LoginCh
	}
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		t, err := template.New("login").Parse(loginTmpl + layoutTmpl)
		if err != nil {
			notify.Warning("ERROR!", err.Error())
			log.Println("error:")
			log.Println(err)
			return
		}
		t.Execute(w, auth.Email)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req.ParseForm()
	auth.Email = req.FormValue("email")
	auth.Password = req.FormValue("password")

	auth.Token = auth.TokenResponse{}
	auth.GetAuth()
	if auth.Token.AccessToken != "" {
		err := keyring.Set("SugarMateReader", auth.Email, auth.Password)
		if err != nil {
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("EMAIL", auth.Email)
		go func() {
			auth.LoginCh <- true
		}()
		t, err := template.New("Logged In").Parse(closeTmpl)
		if err != nil {
			notify.Warning("ERROR!", err.Error())
			log.Println("error:")
			log.Println(err)
			return
		}
		t.Execute(w, nil)
		return
	}
	req.Method = http.MethodGet
	handleLogin(w, req)
}

func handleSettings(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.Method == http.MethodPost {
		req.ParseForm()
	}
	t, err := template.New("settings").Parse(settingsTmpl + layoutTmpl)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
		return
	}
	t.Execute(w, auth.Email)
	return
}

// OpenLogin will open the login window
func OpenLogin() {
	browser.OpenURL(url + "/login")
}

// OpenSettings will open the settings window
func OpenSettings() {
	browser.OpenURL(url + "/settings")
}
