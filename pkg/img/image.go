package img

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/style"
	"github.com/brettcodling/SugarMateReader/pkg/database"
	"github.com/brettcodling/SugarMateReader/pkg/directory"
	"github.com/brettcodling/SugarMateReader/pkg/notify"
	"github.com/fogleman/gg"
)

var (
	lowAlertLevel                               float64
	highAlertLevel                              float64
	lowRangeLevel                               float64
	highRangeLevel                              float64
	lowRangeLevelField, highRangeLevelField     nucular.TextEditor
	lowAlertLevelField, highAlertLevelField     nucular.TextEditor
	lowAlertEnabledField, highAlertEnabledField bool
	lowAlertEnabled, highAlertEnabled           bool
	SettingsCh                                  chan struct{}
	missingLastDelta                            bool
	missingLastTrend                            bool
	missingLastValue                            bool
	highLastValue                               bool
)

// init initialises the image environment variables.
func init() {
	lowAlertEnabledString := database.Get("LOW_ALERT_ENABLED")
	if lowAlertEnabledString != "" {
		lowAlertEnabled, _ = strconv.ParseBool(lowAlertEnabledString)
	}
	highAlertEnabledString := database.Get("HIGH_ALERT_ENABLED")
	if highAlertEnabledString != "" {
		highAlertEnabled, _ = strconv.ParseBool(highAlertEnabledString)
	}
	lowAlertString := database.Get("LOW_ALERT")
	if lowAlertString != "" {
		lowAlertLevel, _ = strconv.ParseFloat(lowAlertString, 64)
	}
	highAlertString := database.Get("HIGH_ALERT")
	if highAlertString != "" {
		highAlertLevel, _ = strconv.ParseFloat(highAlertString, 64)
	}
	lowRangeString := database.Get("LOW_RANGE")
	if lowRangeString != "" {
		lowRangeLevel, _ = strconv.ParseFloat(lowRangeString, 64)
	} else {
		lowRangeLevel = 4
	}
	highRangeString := database.Get("HIGH_RANGE")
	if highRangeString != "" {
		highRangeLevel, _ = strconv.ParseFloat(highRangeString, 64)
	} else {
		highRangeLevel = 10
	}

	lowRangeLevelField.Flags = nucular.EditField
	lowRangeLevelField.SingleLine = true
	lowRangeLevelField.Maxlen = 4
	highRangeLevelField.Flags = nucular.EditField
	highRangeLevelField.SingleLine = true
	highRangeLevelField.Maxlen = 4
	lowAlertLevelField.Flags = nucular.EditField
	lowAlertLevelField.SingleLine = true
	lowAlertLevelField.Maxlen = 4
	highAlertLevelField.Flags = nucular.EditField
	highAlertLevelField.SingleLine = true
	highAlertLevelField.Maxlen = 4

	SettingsCh = make(chan struct{})
}

// BuildImage builds the entire reading image which is used as the systray icon.
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
		font = directory.Dir + "/assets/NotoSansSymbols.ttf"
	default:
		font = directory.Dir + "/assets/Roboto-Bold.ttf"
	}
	context.LoadFontFace(font, fontSize)
	context.DrawStringAnchored(value, 30, 25, 0.5, 0.5)

	return context
}

// getImageDelta gets the delta image.
func getImageDelta(html string) (image.Image, error) {
	start := `<div class="delta ">`
	index := strings.Index(html, start)
	if index < 0 {
		if !missingLastDelta {
			missingLastDelta = true
			return nil, errors.New("Delta not found.")
		}
		return nil, nil
	}
	missingLastDelta = false
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

// getImageTrend gets the trend image.
func getImageTrend(html string) (image.Image, error) {
	start := `<div class="trend">`
	index := strings.Index(html, start)
	if index < 0 {
		if !missingLastTrend {
			missingLastTrend = true
			return nil, errors.New("Trend not found.")
		}
		return nil, nil
	}
	missingLastTrend = false
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

// getImageValue gets the value image.
func getImageValue(html string) (image.Image, error) {
	start := `<div class="value">`
	index := strings.Index(html, start)
	if index < 0 {
		if !missingLastValue {
			missingLastValue = true
			return nil, errors.New("Value not found.")
		}
		return nil, nil
	}
	value := html[index+len(start):]
	value = value[:strings.Index(value, "</div>")]

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	if lowAlertEnabled && lowAlertLevel > 0 && floatValue <= lowAlertLevel {
		notify.Warning("ALERT!", "LOW GLUCOSE")
	}
	if highAlertEnabled && highAlertLevel > 0 && floatValue >= highAlertLevel {
		if !highLastValue {
			highLastValue = true
			notify.Warning("ALERT!", "HIGH GLUCOSE")
		}
	} else {
		highLastValue = false
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

// OpenSettings will open the settings window
func OpenSettings() {
	lowRangeLevelField.SelectAll()
	lowRangeLevelField.Text([]rune(fmt.Sprintf("%.1f", lowRangeLevel)))
	highRangeLevelField.SelectAll()
	highRangeLevelField.Text([]rune(fmt.Sprintf("%.1f", highRangeLevel)))
	lowAlertLevelField.SelectAll()
	lowAlertLevelField.Text([]rune(fmt.Sprintf("%.1f", lowAlertLevel)))
	highAlertLevelField.SelectAll()
	highAlertLevelField.Text([]rune(fmt.Sprintf("%.1f", highAlertLevel)))
	lowAlertEnabledField = lowAlertEnabled
	highAlertEnabledField = highAlertEnabled
	wnd := nucular.NewMasterWindow(0, "Settings", updateSettings)
	wnd.SetStyle(style.FromTheme(style.DarkTheme, 2.0))
	wnd.Main()
}

// updateSettings will setup the login window and wait for updates
// when the login button is clicked the email will be stored in a file and the password in keyring
func updateSettings(w *nucular.Window) {
	w.Row(50).Dynamic(1)
	w.Label("Range:", "LC")
	w.Row(25).Dynamic(2)
	w.Label("Low", "LC")
	w.Label("High", "LC")
	w.Row(50).Dynamic(2)
	lowRangeLevelField.Edit(w)
	highRangeLevelField.Edit(w)
	w.Row(50).Dynamic(1)
	w.Label("Alerts:", "LC")
	w.Row(25).Dynamic(2)
	w.CheckboxText("Low", &lowAlertEnabledField)
	w.CheckboxText("High", &highAlertEnabledField)
	w.Row(50).Dynamic(2)
	lowAlertLevelField.Edit(w)
	highAlertLevelField.Edit(w)
	w.Row(50).Dynamic(1)
	if w.ButtonText("Save") {
		var err error
		lowRangeString := string(lowRangeLevelField.Buffer)
		lowRangeLevel, err = strconv.ParseFloat(lowRangeString, 64)
		if err != nil {
			log.Println(err)
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("LOW_RANGE", lowRangeString)
		highRangeString := string(highRangeLevelField.Buffer)
		highRangeLevel, err = strconv.ParseFloat(highRangeString, 64)
		if err != nil {
			log.Println(err)
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("HIGH_RANGE", highRangeString)
		lowAlertString := string(lowAlertLevelField.Buffer)
		lowAlertLevel, err = strconv.ParseFloat(lowAlertString, 64)
		if err != nil {
			log.Println(err)
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("LOW_ALERT", lowAlertString)
		highAlertString := string(highAlertLevelField.Buffer)
		highAlertLevel, err = strconv.ParseFloat(highAlertString, 64)
		if err != nil {
			log.Println(err)
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("HIGH_ALERT", highAlertString)
		lowAlertEnabled = lowAlertEnabledField
		lowAlertEnabledString := "false"
		if lowAlertEnabled {
			lowAlertEnabledString = "true"
		}
		database.Set("LOW_ALERT_ENABLED", lowAlertEnabledString)
		highAlertEnabled = highAlertEnabledField
		highAlertEnabledString := "false"
		if highAlertEnabled {
			highAlertEnabledString = "true"
		}
		database.Set("HIGH_ALERT_ENABLED", highAlertEnabledString)
		w.Master().Close()
		SettingsCh <- struct{}{}
	}
}
