package img

import (
	"bytes"
	"errors"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brettcodling/SugarMateReader/pkg/notify"
	"github.com/fogleman/gg"
)

var (
	dir            string
	lowAlertLevel  float64
	highAlertLevel float64
	lowRangeLevel  float64
	highRangeLevel float64
)

func init() {
	path, _ := os.Executable()
	dir = filepath.Dir(path)

	lowAlertString := os.Getenv("LOW_ALERT")
	if lowAlertString != "" {
		lowAlertLevel, _ = strconv.ParseFloat(lowAlertString, 64)
	}
	highAlertString := os.Getenv("HIGH_ALERT")
	if highAlertString != "" {
		highAlertLevel, _ = strconv.ParseFloat(highAlertString, 64)
	}
	lowRangeString := os.Getenv("LOW_RANGE")
	if lowRangeString != "" {
		lowRangeLevel, _ = strconv.ParseFloat(lowRangeString, 64)
	} else {
		lowRangeLevel = 4
	}
	highRangeString := os.Getenv("HIGH_RANGE")
	if highRangeString != "" {
		highRangeLevel, _ = strconv.ParseFloat(highRangeString, 64)
	} else {
		highRangeLevel = 10
	}
}

func BuildImage(html string) []byte {
	fullContext := gg.NewContext(180, 50)
	valueImage, err := getImageValue(html)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
	} else {
		fullContext.DrawImageAnchored(valueImage, 40, 25, 0.5, 0.5)
	}
	trendImage, err := getImageTrend(html)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
	} else {
		fullContext.DrawImageAnchored(trendImage, 90, 25, 0.5, 0.5)
	}
	deltaImage, err := getImageDelta(html)
	if err != nil {
		notify.Warning("ERROR!", err.Error())
		log.Println("error:")
		log.Println(err)
	} else {
		fullContext.DrawImageAnchored(deltaImage, 140, 25, 0.5, 0.5)
	}
	buf := new(bytes.Buffer)
	fullContext.EncodePNG(buf)

	return buf.Bytes()
}

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
		font = dir + "/assets/NotoSansSymbols.ttf"
	default:
		font = dir + "/assets/Roboto-Bold.ttf"
	}
	context.LoadFontFace(font, fontSize)
	context.DrawStringAnchored(value, 30, 25, 0.5, 0.5)

	return context
}

func getImageDelta(html string) (image.Image, error) {
	start := `<div class="delta ">`
	index := strings.Index(html, start)
	if index < 0 {
		return nil, errors.New("Delta not found.")
	}
	delta := html[index+len(start):]
	delta = delta[:strings.Index(delta, "</div>")]

	change, _ := strconv.ParseFloat(delta, 64)
	change = math.Abs(change)
	red := 1.0
	green := 1.0
	blue := 1.0
	if change >= 0.5 {
		green = 0
		blue = 0
	}

	context := getImageContext(delta, "roboto", 26, red, green, blue)
	buf := new(bytes.Buffer)
	context.EncodePNG(buf)
	deltaImage, _, err := image.Decode(buf)
	if err != nil {
		return nil, err
	}

	return deltaImage, nil
}

func getImageTrend(html string) (image.Image, error) {
	start := `<div class="trend">`
	index := strings.Index(html, start)
	if index < 0 {
		return nil, errors.New("Trend not found.")
	}
	trend := html[index+len(start):]
	trend = strings.TrimSpace(trend[:strings.Index(trend, "</div>")])
	trend = trend[10:]
	trend = trend[:len(trend)-4]
	switch true {
	case strings.Contains(trend, "FORTY_FIVE_UP"):
		trend = "↗"
	case strings.Contains(trend, "DOUBLE_UP"):
		trend = "↑↑"
		notify.Warning("ALERT!", "RISING FAST")
	case strings.Contains(trend, "UP"):
		trend = "↑"
	case strings.Contains(trend, "FORTY_FIVE_DOWN"):
		trend = "↘"
	case strings.Contains(trend, "DOUBLE_DOWN"):
		trend = "↓↓"
		notify.Warning("ALERT!", "FALLING FAST")
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

func getImageValue(html string) (image.Image, error) {
	start := `<div class="value">`
	index := strings.Index(html, start)
	if index < 0 {
		return nil, errors.New("Value not found.")
	}
	value := html[index+len(start):]
	value = value[:strings.Index(value, "</div>")]

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	if lowAlertLevel > 0 && floatValue <= lowAlertLevel {
		notify.Warning("ALERT!", "LOW GLUCOSE")
	} else if highAlertLevel > 0 && floatValue >= highAlertLevel {
		notify.Warning("ALERT!", "HIGH GLUCOSE")
	}

	red := 0.0
	green := 1.0
	blue := 0.0
	if floatValue < lowRangeLevel {
		green = 0
		red = 1
	} else if floatValue >= highRangeLevel {
		green = 0.5
		red = 1
	}

	context := getImageContext(value, "roboto", 32, red, green, blue)
	buf := new(bytes.Buffer)
	context.EncodePNG(buf)
	valueImage, _, err := image.Decode(buf)
	if err != nil {
		return nil, err
	}

	return valueImage, nil
}
