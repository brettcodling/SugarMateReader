package main

import (
	"io"
	"log"
	"log/syslog"
	"net/http"
	"time"

	"github.com/brettcodling/SugarMateReader/pkg/auth"
	"github.com/brettcodling/SugarMateReader/pkg/img"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
	"github.com/brettcodling/SugarMateReader/pkg/systray"
)

func main() {
	syslog, err := syslog.New(syslog.LOG_INFO, "SugarMateReader")
	if err != nil {
		log.Fatal("Unable to connect to syslog")
	}
	log.SetOutput(syslog)

	systray.Run(func() {
		quit := systray.AddMenuItem("Quit", "")
		go func() {
			for {
				select {
				case <-quit.ClickedCh:
					systray.Quit()
				}
			}
		}()
		getReading(true)
		tick := time.Tick(1 * time.Minute)
		for {
			select {
			case <-tick:
				getReading(true)
			}
		}
	}, func() {})
}

func getReading(retry bool) {
	req, err := http.NewRequest(http.MethodGet, "https://sugarmate.io/nightstand", nil)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: auth.Token.AccessToken,
	})
	transport := &http.Transport{}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return
	}
	if resp.StatusCode != http.StatusOK {
		if retry {
			auth.GetAuth()
			getReading(false)
		}

		return
	}
	systray.SetIcon(img.BuildImage(string(body)))
}
