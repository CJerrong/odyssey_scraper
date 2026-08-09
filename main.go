package main

import (
	"log"
	"time"

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

	state, err := loadState()
	if err != nil {
		log.Fatalf("failed to load state: %v", err)
	}

	targetDate := getTargetSunday(state)

	log.Printf("Checking at %s\n", time.Now().UTC().Format(time.RFC3339))
	log.Println("Last notified:", state.LastNotifiedSunday)
	log.Println("Target Sunday:", targetDate)

	movies, err := fetchShowtimes(targetDate)
	if err != nil {
		log.Fatalf("failed to fetch showtimes: %v", err)
	}

	if !ticketsReleased(movies) {
		log.Println("No tickets released yet.")
		return
	}

	err = sendTelegram(targetDate, movies)
	if err != nil {
		log.Fatalf("failed to send telegram message: %v", err)
	}

	log.Println("Telegram notification sent!")

	state.LastNotifiedSunday = targetDate
	if err := saveState(state); err != nil {
		log.Fatalf("failed to save state (update available): %v", err)
	}

	log.Println("Updated state.json")
}
