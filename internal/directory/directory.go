package directory

import (
	"log"
	"os"
	"path/filepath"
)

var (
	Dir       string
	ConfigDir string
)

func init() {
	path, _ := os.Executable()
	Dir = filepath.Dir(path)

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(configDir+"/SugarMateReader", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	ConfigDir = configDir + "/SugarMateReader/"
}
