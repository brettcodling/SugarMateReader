package notify

import (
	"strconv"

	"github.com/brettcodling/SugarMateReader/internal/directory"
	"github.com/gen2brain/beeep"
)

var (
	highLastValue bool
)

// Warning creates a warning notification.
func Warning(title, context string) {
	beeep.Notify(title, context, directory.ConfigDir+"warning.png")
}

func AlertLow(enabled bool, floatValue float64, lowLevel string) error {
	if enabled {
		lowAlertLevel, err := strconv.ParseFloat(lowLevel, 64)
		if err != nil {
			return err
		}
		if lowAlertLevel > 0 && floatValue <= lowAlertLevel {
			Warning("ALERT!", "LOW GLUCOSE")
		}
	}

	return nil
}

func AlertHigh(enabled bool, floatValue float64, highLevel string) error {
	if enabled {
		if !highLastValue {
			highAlertLevel, err := strconv.ParseFloat(highLevel, 64)
			if err != nil {
				return err
			}
			if highAlertLevel > 0 && floatValue >= highAlertLevel {
				highLastValue = true
				Warning("ALERT!", "HIGH GLUCOSE")
			} else {
				highLastValue = false
			}
		}
	} else {
		highLastValue = false
	}

	return nil
}
