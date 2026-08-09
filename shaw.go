package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func ticketsReleased(movies []Movie) bool {
	return len(filterIMAXShowtimes(movies)) > 0
}

func fetchShowtimes(date string) ([]Movie, error) {
	params := url.Values{}
	params.Set("date", date)
	params.Set("movieId", movieID)
	params.Set("locationId", locationID)
	params.Set("promotionId", promotionID)

	req, err := http.NewRequest(
		http.MethodGet,
		baseURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer",
		fmt.Sprintf(
			"https://shaw.sg/showtimes?movie=The+Odyssey_%s&location=&date=%s",
			movieID,
			date,
		),
	)
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36",
	)
	req.Header.Set("X-App", "PWSM")
	req.Header.Set("X-Api-Forward-To", "internal")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var movies []Movie

	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, err
	}

	return movies, nil
}

func filterIMAXShowtimes(movies []Movie) []ShowTime {
	var result []ShowTime

	for _, movie := range movies {
		for _, show := range movie.ShowTimes {
			if show.LocationVenueBrandCode == "IMAX" {
				result = append(result, show)
			}
		}
	}

	return result
}