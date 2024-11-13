package main

import (
	"log"
	"log/syslog"
	"os"
	"time"

	"github.com/brettcodling/SugarMateReader/pkg/auth"
	"github.com/brettcodling/SugarMateReader/pkg/database"
	"github.com/brettcodling/SugarMateReader/pkg/img"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
	"github.com/brettcodling/SugarMateReader/pkg/readings"
	"github.com/brettcodling/SugarMateReader/pkg/systray"
	"github.com/go-co-op/gocron"
	"github.com/pkg/browser"
)

func main() {
	defer database.DB.Close()
	if os.Getenv("DISABLE_SYSLOG") != "1" {
		syslog, err := syslog.New(syslog.LOG_INFO, "SugarMateReader")
		if err != nil {
			log.Fatal("Unable to connect to syslog")
		}
		log.SetOutput(syslog)
	}

	systray.Run(func() {
		goToUrl := systray.AddMenuItem("Go To Nightstand", "")
		login := systray.AddMenuItem("Login", "")
		settings := systray.AddMenuItem("Settings", "")
		systray.AddSeparator()
		quit := systray.AddMenuItem("Quit", "")
		go func() {
			for {
				select {
				case <-goToUrl.ClickedCh:
					browser.OpenURL("https://sugarmate.io/nightstand")
				case <-login.ClickedCh:
					go auth.OpenLogin()
				case <-settings.ClickedCh:
					go img.OpenSettings()
				case <-quit.ClickedCh:
					systray.Quit()
				case <-auth.LoginCh:
					setIcon()
				case <-img.SettingsCh:
					setIcon()
				}
			}
		}()
		setIcon()
		if readings.LastUpdateTime == "" {
			notify.Warning("ERROR!", "Couldn't get any readings")
			log.Fatal("No readings available")
		}
		tz, _ := time.LoadLocation("Local")
		if tz == nil {
			tz = time.UTC
		}
		s := gocron.NewScheduler(tz)
		specificTime, err := time.Parse(time.RFC3339Nano, readings.LastUpdateTime)
		if err != nil {
			notify.Warning("ERROR!", "Failed to parse last reading time")
			log.Fatal("Failed to parse last reading time")
		}
		s.Every(5).Minutes().StartAt(specificTime.Add(10 * time.Second)).Do(setIcon)
	}, func() {})
}

// setIcon sets the systray icon to the reading image.
func setIcon() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			notify.Warning("ERROR!", "Failed to set reading")
		}
	}()
	reading := readings.GetReading(true, "", "")
	if len(reading) > 0 {
		systray.SetIcon(reading)
	}
}
