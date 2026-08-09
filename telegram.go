package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type telegramRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func sendTelegram(targetDate string, movies []Movie) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}

	chatIDs := os.Getenv("TELEGRAM_CHAT_IDS")
	if chatIDs == "" {
		return fmt.Errorf("TELEGRAM_CHAT_IDS is not set")
	}

	message := buildTelegramMessage(targetDate, movies)

	var failed []string

	for _, chatID := range strings.Split(chatIDs, ",") {
		chatID = strings.TrimSpace(chatID)

		if err := sendMessage(token, chatID, message); err != nil {
			log.Printf("Failed to send to chat %s: %v", chatID, err)
			failed = append(failed, chatID)
			continue
		}

		log.Printf("Successfully sent notification to chat %s", chatID)
	}

	if len(failed) > 0 {
		return fmt.Errorf(
			"failed to send notification to %d chat(s): %s",
			len(failed),
			strings.Join(failed, ", "),
		)
	}

	return nil
}

func sendMessage(token, chatID, message string) error {

	body := telegramRequest{
		ChatID: chatID,
		Text:   message,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		token,
	)

	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned %s", resp.Status)
	}

	return nil
}

func buildTelegramMessage(targetDate string, movies []Movie) string {

	imaxShows := filterIMAXShowtimes(movies)

	grouped := make(map[string][]string)

	for _, show := range imaxShows {
		grouped[show.Venue] = append(grouped[show.Venue], show.DisplayTime)
	}

	// Sort venue names alphabetically.
	// Put Lido at the top of the list if it exists
	var venues []string

	if _, ok := grouped["Lido IMAX"]; ok {
		venues = append(venues, "Lido IMAX")
	}
	
	start := 0
	if len(venues) > 0 && venues[0] == "Lido IMAX" {
		start = 1
	}

	for venue := range grouped {
		if venue != "Lido IMAX" {
			venues = append(venues, venue)
		}
	}

	sort.Strings(venues[start:])

	var sb strings.Builder

	sb.WriteString("🎬 The Odyssey IMAX tickets are now available!\n\n")

	if parsedDate, err := time.Parse("2006-01-02", targetDate); err == nil {
		sb.WriteString(fmt.Sprintf(
			"📅 %s\n",
			parsedDate.Format("Monday 02 Jan 2006"),
		))
	} else {
		sb.WriteString(fmt.Sprintf("📅 %s\n", targetDate))
	}

	sb.WriteString(fmt.Sprintf(
		"🎟️ %d IMAX showtimes found\n\n",
		len(imaxShows),
	))

	sb.WriteString("Showtimes:\n\n")

	for _, venue := range venues {

		sb.WriteString(fmt.Sprintf("• %s\n", venue))
		sb.WriteString("  ")
		sb.WriteString(strings.Join(grouped[venue], ", "))
		sb.WriteString("\n\n")
	}

	sb.WriteString("Book now:\n")

	sb.WriteString(fmt.Sprintf(
		"https://shaw.sg/showtimes?movie=The+Odyssey_%s&location=&date=%s",
		movieID,
		targetDate,
	))

	return sb.String()
}