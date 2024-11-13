package img

import (
	"bytes"
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
	fastChange                                  float64
	lowAlertLevel                               float64
	highAlertLevel                              float64
	lowRangeLevel                               float64
	highRangeLevel                              float64
	fastChangeField                             nucular.TextEditor
	lowRangeLevelField, highRangeLevelField     nucular.TextEditor
	lowAlertLevelField, highAlertLevelField     nucular.TextEditor
	lowAlertEnabledField, highAlertEnabledField bool
	lowAlertEnabled, highAlertEnabled           bool
	SettingsCh                                  chan struct{}
	missingLastDelta                            bool
	missingLastTrend                            bool
	missingLastValue                            bool
	highLastValue                               bool
	lastunit                                    string
	unit                                        string
	format                                      = "%.1f"
)

// init initialises the image environment variables.
func init() {
	unit = database.Get("UNIT")
	if unit == "" {
		unit = "mmol"
	}
	if unit == "mgdl" {
		format = "%.0f"
	}
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
	fastChangeString := database.Get("FAST_CHANGE")
	if fastChangeString != "" {
		fastChange, _ = strconv.ParseFloat(fastChangeString, 64)
	} else {
		fastChange = 0.5
	}

	lowRangeLevelField.Flags = nucular.EditField
	lowRangeLevelField.SingleLine = true
	lowRangeLevelField.Maxlen = 5
	highRangeLevelField.Flags = nucular.EditField
	highRangeLevelField.SingleLine = true
	highRangeLevelField.Maxlen = 5
	lowAlertLevelField.Flags = nucular.EditField
	lowAlertLevelField.SingleLine = true
	lowAlertLevelField.Maxlen = 5
	highAlertLevelField.Flags = nucular.EditField
	highAlertLevelField.SingleLine = true
	highAlertLevelField.Maxlen = 5
	fastChangeField.Flags = nucular.EditField
	fastChangeField.SingleLine = true
	fastChangeField.Maxlen = 5

	SettingsCh = make(chan struct{})
}

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
		font = directory.Dir + "/assets/NotoSansSymbols.ttf"
	default:
		font = directory.Dir + "/assets/Roboto-Bold.ttf"
	}
	context.LoadFontFace(font, fontSize)
	context.DrawStringAnchored(value, 30, 25, 0.5, 0.5)

	return context
}

// getImageDelta gets the delta image.
func getImageDelta(delta int) (image.Image, error) {
	change := float64(delta)
	if unit == "mmol" {
		change = change / 18
	}
	red := 1.0
	green := 1.0
	blue := 1.0
	if math.Abs(change) >= fastChange {
		green = 0
		blue = 0
	}

	context := getImageContext(fmt.Sprintf(format, change), "roboto", 26, red, green, blue)
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
func getImageValue(value int) (image.Image, error) {
	floatValue := float64(value)
	if unit == "mmol" {
		floatValue = floatValue / 18
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

	context := getImageContext(fmt.Sprintf(format, floatValue), "roboto", 32, red, green, blue)
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
	lastunit = unit
	fastChangeField.SelectAll()
	fastChangeField.Text([]rune(fmt.Sprintf(format, fastChange)))
	lowRangeLevelField.SelectAll()
	lowRangeLevelField.Text([]rune(fmt.Sprintf(format, lowRangeLevel)))
	highRangeLevelField.SelectAll()
	highRangeLevelField.Text([]rune(fmt.Sprintf(format, highRangeLevel)))
	lowAlertLevelField.SelectAll()
	lowAlertLevelField.Text([]rune(fmt.Sprintf(format, lowAlertLevel)))
	highAlertLevelField.SelectAll()
	highAlertLevelField.Text([]rune(fmt.Sprintf(format, highAlertLevel)))
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
	w.Label("Fast Change", "LC")
	fastChangeField.Edit(w)
	w.Row(50).Dynamic(1)
	w.Label("Units:", "LC")
	w.Row(50).Dynamic(2)
	if w.OptionText("mmol/l", lastunit == "mmol") {
		if lastunit == "mgdl" {
			factor := 18.0
			if unit == "mmol" {
				factor = 1.0
			}
			fastChangeField.SelectAll()
			fastChangeField.Text([]rune(fmt.Sprintf("%.1f", fastChange/factor)))
			lowRangeLevelField.SelectAll()
			lowRangeLevelField.Text([]rune(fmt.Sprintf("%.1f", lowRangeLevel/factor)))
			highRangeLevelField.SelectAll()
			highRangeLevelField.Text([]rune(fmt.Sprintf("%.1f", highRangeLevel/factor)))
			lowAlertLevelField.SelectAll()
			lowAlertLevelField.Text([]rune(fmt.Sprintf("%.1f", lowAlertLevel/factor)))
			highAlertLevelField.SelectAll()
			highAlertLevelField.Text([]rune(fmt.Sprintf("%.1f", highAlertLevel/factor)))
		}
		lastunit = "mmol"
	}
	if w.OptionText("mg/dl", lastunit == "mgdl") {
		if lastunit == "mmol" {
			factor := 18.0
			if unit == "mgdl" {
				factor = 1.0
			}
			fastChangeField.SelectAll()
			fastChangeField.Text([]rune(fmt.Sprintf("%.0f", fastChange*factor)))
			lowRangeLevelField.SelectAll()
			lowRangeLevelField.Text([]rune(fmt.Sprintf("%.0f", lowRangeLevel*factor)))
			highRangeLevelField.SelectAll()
			highRangeLevelField.Text([]rune(fmt.Sprintf("%.0f", highRangeLevel*factor)))
			lowAlertLevelField.SelectAll()
			lowAlertLevelField.Text([]rune(fmt.Sprintf("%.0f", lowAlertLevel*factor)))
			highAlertLevelField.SelectAll()
			highAlertLevelField.Text([]rune(fmt.Sprintf("%.0f", highAlertLevel*factor)))
		}
		lastunit = "mgdl"
	}
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
		fastChangeString := string(fastChangeField.Buffer)
		fastChange, err = strconv.ParseFloat(fastChangeString, 64)
		if err != nil {
			log.Println(err)
			notify.Warning("ERROR!", err.Error())
		}
		database.Set("FAST_CHANGE", fastChangeString)
		database.Set("UNIT", lastunit)
		unit = lastunit
		if unit == "mmol" {
			format = "%.1f"
		}
		if unit == "mgdl" {
			format = "%.0f"
		}
		w.Master().Close()
		SettingsCh <- struct{}{}
	}
}
