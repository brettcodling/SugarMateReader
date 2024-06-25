package main

import (
	"log"
	"log/syslog"
	"time"

	"github.com/brettcodling/SugarMateReader/pkg/readings"
	"github.com/brettcodling/SugarMateReader/pkg/systray"
	"github.com/pkg/browser"
)

func main() {
	syslog, err := syslog.New(syslog.LOG_INFO, "SugarMateReader")
	if err != nil {
		log.Fatal("Unable to connect to syslog")
	}
	log.SetOutput(syslog)

	systray.Run(func() {
		refresh := systray.AddMenuItem("Refresh", "")
		goToUrl := systray.AddMenuItem("Go To Nightstand", "")
		systray.AddSeparator()
		quit := systray.AddMenuItem("Quit", "")
		go func() {
			for {
				select {
				case <-refresh.ClickedCh:
					setIcon()
				case <-goToUrl.ClickedCh:
					browser.OpenURL("https://sugarmate.io/nightstand")
				case <-quit.ClickedCh:
					systray.Quit()
				}
			}
		}()
		setIcon()
		tick := time.Tick(1 * time.Minute)
		for {
			select {
			case <-tick:
				setIcon()
			}
		}
	}, func() {})
}

// setIcon sets the systray icon to the reading image.
func setIcon() {
	reading := readings.GetReading(true)
	if len(reading) > 0 {
		systray.SetIcon(reading)
	}
}
