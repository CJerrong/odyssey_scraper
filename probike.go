package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const probikeURL = "https://www.probike.com.sg/products/x-lab-rs5"

type BikeVariant struct {
	ID        string
	Name      string
	Available bool
}

type ProductJSONLD struct {
	Name   string `json:"name"`
	Offers []struct {
		Name         string `json:"name"`
		Availability string `json:"availability"`
		URL          string `json:"url"`
	} `json:"offers"`
}

func runProBikeMonitor() error {
	log.Println("Checking ProBike stock...")

	variants, err := getProBikeVariants()
	if err != nil {
		return err
	}

	state, err := loadState()
	if err != nil {
		return err
	}

	var newlyAvailable []BikeVariant

	for _, variant := range variants {
		log.Printf(
			"%s: available=%t",
			variant.Name,
			variant.Available,
		)

		if variant.Available {
			if !containsVariant(
				state.ProBike.NotifiedVariants,
				variant.ID,
			) {
				newlyAvailable = append(
					newlyAvailable,
					variant,
				)

				state.ProBike.NotifiedVariants =
					append(
						state.ProBike.NotifiedVariants,
						variant.ID,
					)
			}
		} else {
			// If bike is not available, remove it from notified list
			state.ProBike.NotifiedVariants =
				removeVariant(
					state.ProBike.NotifiedVariants,
					variant.ID,
				)
		}
	}

	if len(newlyAvailable) > 0 {
		if err := sendProbikeTelegram(newlyAvailable); err != nil {
			return fmt.Errorf(
				"failed to send ProBike notification: %w",
				err,
			)
		}

		log.Printf(
			"Notified about %d newly available variant(s).",
			len(newlyAvailable),
		)
	}

	return saveState(state)
}

func getProBikeVariants() ([]BikeVariant, error) {
	resp, err := http.Get(probikeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ProBike page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ProBike returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ProBike HTML: %w", err)
	}

	var variants []BikeVariant

	doc.Find("script[type='application/ld+json']").EachWithBreak(
		func(_ int, s *goquery.Selection) bool {
			var product ProductJSONLD

			if err := json.Unmarshal(
				[]byte(strings.TrimSpace(s.Text())),
				&product,
			); err != nil {
				return true
			}

			if product.Name != "X-LAB RS5" {
				return true
			}

			for _, offer := range product.Offers {
				if !strings.HasSuffix(offer.Name, " / S") {
					continue
				}

				variants = append(variants, BikeVariant{
					Name:      offer.Name,
					Available: strings.Contains(offer.Availability, "InStock"),
				})
			}

			return false
		},
	)

	if len(variants) == 0 {
		return nil, fmt.Errorf("could not find any size S variants")
	}

	return variants, nil
}