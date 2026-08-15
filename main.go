package main

import (
	"log"

	"github.com/joho/godotenv"
)

const (
	baseURL     = "https://shaw.sg/internal/get_show_times"
	movieID     = "1238"
	locationID  = "0"
	promotionID = "0"
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using environment variables")
    }

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if config.Monitors.Odyssey.Enabled {
		log.Println("Running Odyssey monitor...")

		err := runOdysseyMonitor()
		if err != nil {
			log.Printf("Odyssey monitor failed: %v", err)
		}
	}

	if config.Monitors.ProBike.Enabled {
		log.Println("Running ProBike monitor...")

		err := runProBikeMonitor()
		if err != nil {
			log.Printf("ProBike monitor failed: %v", err)
		}
	}

	return
}
