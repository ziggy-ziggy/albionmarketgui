package client

import (
	"fmt"
	"os"

	"github.com/broderickhyman/albiondata-client/log"
)

var version string

//Client struct base
type Client struct {
}

//NewClient return a new Client instance
func NewClient(_version string) *Client {
	version = _version
	return &Client{}
}

//Run starts client settings and run
func (client *Client) Run() error {
	log.Infof("Starting Albion Data Client, version: %s", version)
	log.Info("wwwwThis is a third-party application and is in no way affiliated with Sandbox Interactive or Albion Online.")
	log.Info("Additional parameters can listed by calling this file with the -h parameter.")

	file, err := os.Open("marketdatastorage/file.txt")
	if err != nil {
		fmt.Println("opening file error", err)
	}
	defer file.Close()

	// read our opened jsonFile as a byte array.
	//byteValue, _ := ioutil.ReadAll(file)

	path := "marketdatastorage"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModeDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	/*if err := os.WriteFile("marketdatastorage/fileFormatted.txt", []byte(lines[0]), 0666); err != nil {
		log.Fatal(err)
	}*/

	ConfigGlobal.setupDebugEvents()
	ConfigGlobal.setupDebugOperations()

	createDispatcher()

	if ConfigGlobal.Offline {
		processOffline(ConfigGlobal.OfflinePath)
	} else {
		apw := newAlbionProcessWatcher()
		return apw.run()
	}
	return nil
}
