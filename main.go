package main

import (
	"embed"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"
	"time"

	"github.com/brettcodling/SugarMateReader/internal/auth"
	"github.com/brettcodling/SugarMateReader/internal/database"
	"github.com/brettcodling/SugarMateReader/internal/directory"
	"github.com/brettcodling/SugarMateReader/internal/notify"
	"github.com/brettcodling/SugarMateReader/internal/readings"
	"github.com/brettcodling/SugarMateReader/internal/ui"
	"github.com/getlantern/systray"
	"github.com/go-co-op/gocron"
	"github.com/pkg/browser"
)

var (
	//go:embed assets/*
	assets             embed.FS
	lastUpdateMenuItem *systray.MenuItem
)

func init() {
	files, err := assets.ReadDir("assets")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		content, err := assets.ReadFile("assets/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		os.WriteFile(directory.ConfigDir+file.Name(), content, os.ModePerm)
	}
}

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
		lastUpdateMenuItem = systray.AddMenuItem("", "")
		lastUpdateMenuItem.Disable()
		goToUrl := systray.AddMenuItem("Open in browser", "")
		login := systray.AddMenuItem("Login", "")
		if _, err := os.Stat("/usr/local/bin/SugarMateReader"); err == nil {
			checked := false
			if _, err := os.Stat(directory.ConfigDir + "./../autostart/SugarMateReader.desktop"); err == nil {
				checked = true
			}
			autostart := *systray.AddMenuItemCheckbox("Launch on start up", "", checked)
			go func() {
				for range autostart.ClickedCh {
					if autostart.Checked() {
						autostart.Uncheck()
					} else {
						autostart.Check()
					}
					err := setAutostart(autostart.Checked())
					if err != nil {
						log.Println("error:")
						log.Println(err)
						if autostart.Checked() {
							autostart.Check()
						} else {
							autostart.Uncheck()
						}
					}
				}
			}()
		}
		settings := systray.AddMenuItem("Settings", "")
		systray.AddSeparator()
		quit := systray.AddMenuItem("Quit", "")
		go func() {
			for {
				select {
				case <-goToUrl.ClickedCh:
					browser.OpenURL("https://app.sugarmate.io")
				case <-login.ClickedCh:
					go ui.OpenLogin()
				case <-settings.ClickedCh:
					go ui.OpenSettings()
				case <-quit.ClickedCh:
					systray.Quit()
				case <-auth.LoginCh:
					setIcon()
				case <-ui.SettingsCh:
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
		s.StartAsync()
	}, func() {})
}

// setAutostart will configure the application to run on startup
func setAutostart(autostart bool) error {
	if autostart {
		os.WriteFile(directory.ConfigDir+"./../autostart/SugarMateReader.desktop", []byte(strings.TrimSpace(`
			#!/usr/bin/env xdg-open
			[Desktop Entry]
			Terminal=false
			Type=Application
			Name=SugarMateReader
			Exec=SugarMateReader
			Icon=SugarMateReader
		`)), os.ModePerm)
	} else {
		err := os.Remove(directory.ConfigDir + "./../autostart/SugarMateReader.desktop")
		if err != nil {
			notify.Warning("ERROR!", "Failed to remove autostart config.")
			return err
		}
	}

	return nil
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
		lastUpdateTime, err := time.Parse(time.RFC3339Nano, readings.LastUpdateTime)
		if err != nil {
			log.Println(err)
			notify.Warning("ERROR!", "Failed to parse last update time")
			return
		}
		lastUpdateMenuItem.SetTitle(fmt.Sprintf("Last updated: %s", lastUpdateTime.Format(time.TimeOnly)))
	}
}
