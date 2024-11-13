package readings

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/brettcodling/SugarMateReader/pkg/auth"
	"github.com/brettcodling/SugarMateReader/pkg/img"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
)

var LastUpdateTime string

// GetReading gets the reading data from SugarMate and build the systray icon image.
func GetReading(retry bool, before, after string) []byte {
	if before == "" {
		before = time.Now().Format(time.RFC3339Nano)
	}
	if after == "" {
		after = time.Now().Add(-10 * time.Minute).Format(time.RFC3339Nano)
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(
		"https://api.sugarmate.io/api/v3/events?before=%s&after=%s",
		before,
		after,
	), nil)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return []byte{}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token.AccessToken))
	transport := &http.Transport{}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return []byte{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return []byte{}
	}
	if resp.StatusCode != http.StatusOK {
		if retry {
			auth.GetAuth()
			return GetReading(false, before, after)
		}

		return []byte{}
	}
	reading, err := parseReading(body)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return []byte{}
	}
	if reading.MgDl < 1 {
		newAfter, _ := time.Parse(time.RFC3339Nano, before)
		return GetReading(retry, before, newAfter.Add(-3*time.Hour).Format(time.RFC3339Nano))
	}
	log.Printf("%+v\n", reading)
	return img.BuildImage(reading.MgDl, reading.Trend, reading.Delta)
}

type Response struct {
	Events []Event `json:"events"`
}

type Event struct {
	EventType string  `json:"event_type"`
	CreatedAt string  `json:"created_at"`
	Glucose   Glucose `json:"glucose"`
}

type Glucose struct {
	MgDl  int    `json:"mg_dl"`
	Trend string `json:"trend"`
}

type CurrentReading struct {
	MgDl  int
	Trend string
	Delta int
}

func parseReading(body []byte) (CurrentReading, error) {
	var currentReading CurrentReading
	var response Response
	err := json.Unmarshal(body, &response)
	if err != nil {
		return currentReading, err
	}

	events := slices.DeleteFunc(response.Events, func(e Event) bool {
		return e.EventType != "glucose"
	})
	if len(events) < 2 {
		return currentReading, nil
	}

	slices.SortFunc(events, func(a, b Event) int {
		aCreated, _ := time.Parse(time.RFC3339Nano, a.CreatedAt)
		bCreated, _ := time.Parse(time.RFC3339Nano, b.CreatedAt)
		if aCreated.After(bCreated) {
			return 1
		}

		return -1
	})

	LastUpdateTime = events[0].CreatedAt
	currentReading.MgDl = events[0].Glucose.MgDl
	currentReading.Trend = events[0].Glucose.Trend
	currentReading.Delta = currentReading.MgDl - events[1].Glucose.MgDl
	return currentReading, nil
}
