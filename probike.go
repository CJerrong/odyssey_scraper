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

type ProductJSONLD struct {
	Name       string           `json:"name"`
	HasVariant []ProductVariant `json:"hasVariant"`
}

type ProductVariant struct {
	ID     string `json:"@id"`
	Name   string `json:"name"`
	Offers Offer  `json:"offers"`
}

type Offer struct {
	Availability string `json:"availability"`
	URL          string `json:"url"`
}

type BikeVariant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
	URL       string `json:"url"`
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

				state.ProBike.NotifiedVariants = append(
					state.ProBike.NotifiedVariants,
					variant.ID,
				)
			}
		} else {
			// If the variant is no longer available, remove it so that we can notify again
			// if it comes back into stock later
			state.ProBike.NotifiedVariants =
				removeVariant(
					state.ProBike.NotifiedVariants,
					variant.ID,
				)
		}
	}

	if len(newlyAvailable) > 0 {
		err := sendProbikeTelegram(newlyAvailable)
		if err != nil {
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
		return nil, fmt.Errorf(
			"failed to fetch ProBike page: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ProBike returned HTTP %d",
			resp.StatusCode,
		)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse ProBike HTML: %w",
			err,
		)
	}

	var variants []BikeVariant

	doc.Find("script[type='application/ld+json']").Each(
		func(_ int, s *goquery.Selection) {
			var product ProductJSONLD

			if err := json.Unmarshal(
				[]byte(strings.TrimSpace(s.Text())),
				&product,
			); err != nil {
				return
			}

			if product.Name != "X-LAB RS5" {
				return
			}

			for _, variant := range product.HasVariant {
				// Only care about size S.
				if !strings.HasSuffix(variant.Name, " / S") {
					continue
				}

				variants = append(variants, BikeVariant{
					ID:        getVariantID(variant.ID),
					Name:      variant.Name,
					Available: strings.Contains(
						variant.Offers.Availability,
						"InStock",
					),
					URL: variant.Offers.URL,
				})
			}
		},
	)

	if len(variants) == 0 {
		return nil, fmt.Errorf(
			"could not find any size S bike variants",
		)
	}

	return variants, nil
}

func getVariantID(id string) string {
	const prefix = "?variant="
	const suffix = "#variant"

	start := strings.Index(id, prefix)
	if start == -1 {
		return id
	}

	start += len(prefix)

	end := strings.Index(id[start:], suffix)
	if end == -1 {
		return id[start:]
	}

	return id[start : start+end]
}