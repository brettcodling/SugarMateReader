# SugarMateReader [![Go Report Card](https://goreportcard.com/badge/github.com/brettcodling/SugarMateReader)](https://goreportcard.com/report/github.com/brettcodling/SugarMateReader)

# SugarMateReader
A Linux system tray app to display glucose readings from sugarmate.io

## build
```
go get -u ./...
go build .
```

## running
```
EMAIL=... PASSWORD=... ./SugarMateReader
```

## options
There are a few different environment variables you can use to customize your setup
|Variable  |Description                           |Default|Example|
|----------|--------------------------------------|-------|-------|
|LOW_ALERT |Level to trigger a low notification   |None   |4.0    |
|HIGH_ALERT|Level to trigger a high notification  |None   |10.0   |
|LOW_RANGE |Level to display the reading in red   |4.0    |4.0    |
|HIGH_RANGE|Level to display the reading in orange|10.0   |10.0   |

## notes
* By default any deltas greater than or equal to 0.5mmol will display the delta value in red
* By default any double arrow trends will trigger a notification