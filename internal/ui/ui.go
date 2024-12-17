package ui

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"

	_ "embed"

	"github.com/brettcodling/SugarMateReader/internal/auth"
	"github.com/brettcodling/SugarMateReader/internal/database"
	"github.com/brettcodling/SugarMateReader/internal/notify"
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
	Settings     Setting
	SettingsCh   chan bool
	url          string
)

type Setting struct {
	Alerts Alert
	Format string
	Range  Range
	Saved  bool
	Units  string
}

type Alert struct {
	LowEnabled        string
	Low               string
	HighEnabled       string
	High              string
	FastChangeEnabled string
	FastChange        string
}

type Range struct {
	Low  string
	High string
}

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

	Settings.Units = database.Get("UNIT")
	if Settings.Units == "" {
		Settings.Units = "mmol"
	}
	if Settings.Units == "mgdl" {
		Settings.Format = "%.0f"
	} else {
		Settings.Format = "%.1f"
	}
	Settings.Alerts.LowEnabled = database.Get("LOW_ALERT_ENABLED")
	Settings.Alerts.HighEnabled = database.Get("HIGH_ALERT_ENABLED")
	Settings.Alerts.FastChangeEnabled = database.Get("FAST_CHANGE_ENABLED")
	Settings.Alerts.Low = database.Get("LOW_ALERT")
	if Settings.Alerts.Low == "" {
		Settings.Alerts.Low = "4.0"
		if Settings.Units == "mgdl" {
			Settings.Alerts.Low = "72"
		}
	}
	Settings.Alerts.High = database.Get("HIGH_ALERT")
	if Settings.Alerts.High == "" {
		Settings.Alerts.High = "12.0"
		if Settings.Units == "mgdl" {
			Settings.Alerts.High = "216"
		}
	}
	Settings.Range.Low = database.Get("LOW_RANGE")
	if Settings.Range.Low == "" {
		Settings.Range.Low = "4.5"
		if Settings.Units == "mgdl" {
			Settings.Range.Low = "81"
		}
	}
	Settings.Range.High = database.Get("HIGH_RANGE")
	if Settings.Range.High == "" {
		Settings.Range.High = "10.0"
		if Settings.Units == "mgdl" {
			Settings.Range.High = "180"
		}
	}
	Settings.Alerts.FastChange = database.Get("FAST_CHANGE")
	if Settings.Alerts.FastChange == "" {
		Settings.Alerts.FastChange = "0.5"
		if Settings.Units == "mgdl" {
			Settings.Alerts.FastChange = "9"
		}
	}

	SettingsCh = make(chan bool)
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

	saved := false
	if req.Method == http.MethodPost {
		req.ParseForm()
		Settings.Range.Low = req.PostForm["range_low"][0]
		database.Set("LOW_RANGE", req.PostForm["range_low"][0])
		Settings.Range.High = req.PostForm["range_high"][0]
		database.Set("HIGH_RANGE", req.PostForm["range_high"][0])
		_, alertLowEnabled := req.PostForm["alert_low_enabled"]
		if alertLowEnabled {
			Settings.Alerts.LowEnabled = "true"
			database.Set("LOW_ALERT_ENABLED", "true")
			Settings.Alerts.Low = req.PostForm["alert_low"][0]
			database.Set("LOW_ALERT", req.PostForm["alert_low"][0])
		} else {
			Settings.Alerts.LowEnabled = "false"
			database.Set("LOW_ALERT_ENABLED", "false")
			Settings.Alerts.Low = ""
			database.Set("LOW_ALERT", "")
		}
		_, alertHighEnabled := req.PostForm["alert_high_enabled"]
		if alertHighEnabled {
			Settings.Alerts.HighEnabled = "true"
			database.Set("HIGH_ALERT_ENABLED", "true")
			Settings.Alerts.High = req.PostForm["alert_high"][0]
			database.Set("HIGH_ALERT", req.PostForm["alert_high"][0])
		} else {
			Settings.Alerts.HighEnabled = "false"
			database.Set("HIGH_ALERT_ENABLED", "false")
			Settings.Alerts.High = ""
			database.Set("HIGH_ALERT", "")
		}
		_, fastChangeEnabled := req.PostForm["fast_change_enabled"]
		if fastChangeEnabled {
			Settings.Alerts.FastChangeEnabled = "true"
			database.Set("FAST_CHANGE_ENABLED", "true")
		} else {
			Settings.Alerts.FastChangeEnabled = "false"
			database.Set("FAST_CHANGE_ENABLED", "false")
		}
		Settings.Alerts.FastChange = req.PostForm["fast_change"][0]
		database.Set("FAST_CHANGE", req.PostForm["fast_change"][0])
		Settings.Units = req.PostForm["unit"][0]
		database.Set("UNIT", req.PostForm["unit"][0])
		if req.PostForm["unit"][0] == "mmol" {
			Settings.Format = "%.1f"
		} else {
			Settings.Format = "%.0f"
		}
		saved = true
		go func() {
			SettingsCh <- true
		}()
	}
	settings := Settings
	settings.Saved = saved
	t, err := template.New("settings").Parse(settingsTmpl + layoutTmpl)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
		return
	}
	t.Execute(w, settings)
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
