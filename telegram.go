package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func sendTelegram(targetDate string, movies []Movie) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" {
		return fmt.Errorf("missing TELEGRAM_BOT_TOKEN")
	}

	if chatID == "" {
		return fmt.Errorf("missing TELEGRAM_CHAT_ID")
	}

	text := buildTelegramMessage(targetDate, movies)

	body := TelegramMessage{
		ChatID: chatID,
		Text:   text,
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

	var sb strings.Builder

	sb.WriteString("🎬 The Odyssey IMAX tickets are now available!\n")

	parsedDate, _ := time.Parse("2006-01-02", targetDate)
	sb.WriteString(fmt.Sprintf("📅 %s\n", parsedDate.Format("Monday 02 Jan 2006")))
	sb.WriteString(fmt.Sprintf("🎟️ %d IMAX showtimes found\n\n", len(imaxShows)))

	sb.WriteString("First few:\n")

	const maxShows = 5

	for i, show := range imaxShows {
		if i >= maxShows {
			break
		}

		sb.WriteString(fmt.Sprintf(
			"• %s — %s\n",
			show.Venue,
			show.DisplayTime,
		))
	}

	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf(
		"https://shaw.sg/showtimes?movie=The+Odyssey_%s&location=&date=%s",
		movieID,
		targetDate,
	))

	return sb.String()
}

