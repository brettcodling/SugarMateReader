package img

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/brettcodling/SugarMateReader/internal/directory"
	"github.com/brettcodling/SugarMateReader/internal/notify"
	"github.com/brettcodling/SugarMateReader/internal/ui"
	"github.com/fogleman/gg"
)

var (
	highLastValue bool
)

// BuildImage builds the entire reading image which is used as the systray icon.
func BuildImage(value int, trend string, delta int) []byte {
	fullContext := gg.NewContext(180, 50)
	valueImage, err := getImageValue(value)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
	} else {
		if valueImage.Bounds().Max.X > 0 || valueImage.Bounds().Max.Y > 0 {
			fullContext.DrawImageAnchored(valueImage, 40, 25, 0.5, 0.5)
		}
	}
	trendImage, err := getImageTrend(trend)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
	} else {
		if trendImage.Bounds().Max.X > 0 || trendImage.Bounds().Max.Y > 0 {
			fullContext.DrawImageAnchored(trendImage, 90, 25, 0.5, 0.5)
		}
	}
	deltaImage, err := getImageDelta(delta)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
	} else if deltaImage != nil {
		if deltaImage.Bounds().Max.X > 0 || deltaImage.Bounds().Max.Y > 0 {
			fullContext.DrawImageAnchored(deltaImage, 140, 25, 0.5, 0.5)
		}
	}
	buf := new(bytes.Buffer)
	fullContext.EncodePNG(buf)

	return buf.Bytes()
}

// getImageContext gets an image context which can be used to build individual images.
func getImageContext(value, font string, fontSize, red, green, blue float64) *gg.Context {
	context := gg.NewContext(80, 50)
	context.SetRGBA(0, 0, 0, 0)
	context.Clear()
	context.SetRGB(red, green, blue)
	if fontSize == 0 {
		fontSize = 32
	}
	switch font {
	case "noto":
		font = directory.ConfigDir + "NotoSansSymbols.ttf"
	default:
		font = directory.ConfigDir + "Roboto-Bold.ttf"
	}
	context.LoadFontFace(font, fontSize)
	context.DrawStringAnchored(value, 30, 25, 0.5, 0.5)

	return context
}

// getImageDelta gets the delta image.
func getImageDelta(delta int) (image.Image, error) {
	change := float64(delta)
	if ui.Settings.Units == "mmol" {
		change = change / 18
	}
	red := 1.0
	green := 1.0
	blue := 1.0
	fastChange, err := strconv.ParseFloat(ui.Settings.Alerts.FastChange, 64)
	if err != nil {
		return nil, err
	}
	if math.Abs(change) >= fastChange {
		green = 0
		blue = 0
		if ui.Settings.Alerts.FastChangeEnabled == "true" {
			if change > 0 {
				notify.Warning("ALERT!", "RISING FAST")
			} else {
				notify.Warning("ALERT!", "FALLING FAST")
			}
		}
	}

	context := getImageContext(fmt.Sprintf(ui.Settings.Format, change), "roboto", 26, red, green, blue)
	buf := new(bytes.Buffer)
	context.EncodePNG(buf)
	deltaImage, _, err := image.Decode(buf)
	if err != nil {
		return nil, err
	}

	return deltaImage, nil
}

// getImageTrend gets the trend image.
func getImageTrend(trend string) (image.Image, error) {
	switch true {
	case strings.Contains(trend, "FORTY_FIVE_UP"):
		trend = "↗"
	case strings.Contains(trend, "DOUBLE_UP"):
		trend = "↑↑"
	case strings.Contains(trend, "UP"):
		trend = "↑"
	case strings.Contains(trend, "FORTY_FIVE_DOWN"):
		trend = "↘"
	case strings.Contains(trend, "DOUBLE_DOWN"):
		trend = "↓↓"
	case strings.Contains(trend, "DOWN"):
		trend = "↓"
	case strings.Contains(trend, "FLAT"):
		trend = "→"
	default:
		trend = "..."
	}

	context := getImageContext(trend, "noto", 32, 1, 1, 1)
	buf := new(bytes.Buffer)
	context.EncodePNG(buf)
	trendImage, _, err := image.Decode(buf)
	if err != nil {
		return nil, err
	}

	return trendImage, nil
}

// getImageValue gets the value image.
func getImageValue(value int) (image.Image, error) {
	floatValue := float64(value)
	if ui.Settings.Units == "mmol" {
		floatValue = floatValue / 18
	}
	if ui.Settings.Alerts.LowEnabled == "true" {
		lowAlertLevel, err := strconv.ParseFloat(ui.Settings.Alerts.Low, 64)
		if err != nil {
			return nil, err
		}
		if lowAlertLevel > 0 && floatValue <= lowAlertLevel {
			notify.Warning("ALERT!", "LOW GLUCOSE")
		}
	}
	if ui.Settings.Alerts.HighEnabled == "true" {
		if !highLastValue {
			highAlertLevel, err := strconv.ParseFloat(ui.Settings.Alerts.High, 64)
			if err != nil {
				return nil, err
			}
			if highAlertLevel > 0 && floatValue >= highAlertLevel {
				highLastValue = true
				notify.Warning("ALERT!", "HIGH GLUCOSE")
			} else {
				highLastValue = false
			}
		}
	} else {
		highLastValue = false
	}

	red := 0.0
	green := 1.0
	blue := 0.0
	lowRangeLevel, err := strconv.ParseFloat(ui.Settings.Range.Low, 64)
	if err != nil {
		return nil, err
	}
	highRangeLevel, err := strconv.ParseFloat(ui.Settings.Range.High, 64)
	if err != nil {
		return nil, err
	}
	if floatValue < lowRangeLevel {
		green = 0
		red = 1
	} else if floatValue >= highRangeLevel {
		green = 0.5
		red = 1
	}

	context := getImageContext(fmt.Sprintf(ui.Settings.Format, floatValue), "roboto", 32, red, green, blue)
	buf := new(bytes.Buffer)
	context.EncodePNG(buf)
	valueImage, _, err := image.Decode(buf)
	if err != nil {
		return nil, err
	}

	return valueImage, nil
}
