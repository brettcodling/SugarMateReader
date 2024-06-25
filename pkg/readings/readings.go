package readings

import (
	"io"
	"log"
	"net/http"

	"github.com/brettcodling/SugarMateReader/pkg/auth"
	"github.com/brettcodling/SugarMateReader/pkg/img"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
)

// GetReading gets the reading data from SugarMate and build the systray icon image.
func GetReading(retry bool) []byte {
	req, err := http.NewRequest(http.MethodGet, "https://sugarmate.io/nightstand", nil)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)

		return []byte{}
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
			return GetReading(false)
		}

		return []byte{}
	}
	return img.BuildImage(string(body))
}
