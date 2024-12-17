# SugarMateReader [![Go Report Card](https://goreportcard.com/badge/github.com/brettcodling/SugarMateReader)](https://goreportcard.com/report/github.com/brettcodling/SugarMateReader)
A Linux system tray app to display glucose readings from sugarmate.io

## build
```
go get -u ./...
go build .
```

## external requirements
* You must have a sugarmate account linked to a glucose source

## running
```
./SugarMateReader
```

## notes
* https://github.com/getlantern/systray is included in the pkg directory in order to build correctly
