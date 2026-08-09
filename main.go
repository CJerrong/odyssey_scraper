package main

import (
	"fmt"
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

	fmt.Printf("Checking at %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Println("Last notified:", state.LastNotifiedSunday)
	fmt.Println("Target Sunday:", targetDate)

	movies, err := fetchShowtimes(targetDate)
	if err != nil {
		log.Fatalf("failed to fetch showtimes: %v", err)
	}

	if !ticketsReleased(movies) {
		fmt.Println("No tickets released yet.")
		return
	}

	err = sendTelegram(targetDate, movies)
	if err != nil {
		log.Fatalf("failed to send telegram message: %v", err)
	}

	fmt.Println("Telegram notification sent!")

	state.LastNotifiedSunday = targetDate
	if err := saveState(state); err != nil {
		log.Fatalf("failed to save state (update available): %v", err)
	}

	fmt.Println("Updated state.json")
}
