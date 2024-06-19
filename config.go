package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

func loadConfig() {
	fmt.Println("Opening" + configFile)
	_, err := os.Stat(configFile)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Creating " + configFile)
		file, fOpenError := os.Create(configFile)
		if fOpenError != nil {
			log.Fatal(fOpenError)
		}
		fmt.Println("Writing defaults to " + configFile)
		// Write the default config to the file
		json.NewEncoder(file).Encode(Configuration{
			HttpPort:        8080,
			OSCPort:         7001,
			SoundFolder:     "sounds",
			DefaultFadeTime: 3000,
		})
		fmt.Println("Wrote to conf.json")
		file.Close()
	}
	file, fOpenError := os.Open(configFile)
	if fOpenError != nil {
		log.Fatal(fOpenError)
	}

	// Decode the config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}
	file.Sync()
	file.Close()
	// Create /sounds if not exists
	if _, err := os.Stat(conf.SoundFolder); os.IsNotExist(err) {
		fmt.Println("Created sounds folder")
		os.Mkdir(conf.SoundFolder, 0777)
	}

}
func saveConfig(newConfig Configuration) error {
	fmt.Println("saveConfig Start")
	conf = newConfig
	file, _ := os.Create(configFile) // Try to open the file

	err := json.NewEncoder(file).Encode(conf)
	if err != nil {
		fmt.Println(err)
		file.Close()
		return err
	}
	file.Sync()
	file.Close()
	fmt.Println("saveConfig End")
	return nil
}
