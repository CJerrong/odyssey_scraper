package main

import (
	"encoding/json"
	"os"
	"time"
)

type State struct {
	LastNotifiedSunday string `json:"lastNotifiedSunday"`
	ProBike ProBikeState `json:"probike"`
}

type ProBikeState struct {
	NotifiedVariants []string `json:"notifiedVariants"`
}

// Load last sunday notified state
func loadState() (*State, error) {
	data, err := os.ReadFile("state.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var state State

	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save last sunday notified state
func saveState(state *State) error {
	bytes, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile("state.json", bytes, 0644)
}

func getTargetSunday(state *State) string {
	if state.LastNotifiedSunday == "" {
		return getNextSunday(time.Now())
	}

	last, err := time.Parse("2006-01-02", state.LastNotifiedSunday)
	if err != nil {
		panic(err)
	}

	return last.AddDate(0, 0, 7).Format("2006-01-02")
}

func getNextSunday(now time.Time) string {
	days := (7 - int(now.Weekday())) % 7

	if days == 0 {
		days = 7
	}

	return now.AddDate(0, 0, days).Format("2006-01-02")
}

func containsVariant(variants []string, id string) bool {
	for _, variantID := range variants {
		if variantID == id {
			return true
		}
	}

	return false
}

func removeVariant(variants []string, id string) []string {
	result := make([]string, 0, len(variants))

	for _, variantID := range variants {
		if variantID != id {
			result = append(result, variantID)
		}
	}

	return result
}