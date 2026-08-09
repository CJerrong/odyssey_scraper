package main

type Movie struct {
    MovieID      int        `json:"movieId"`
    PrimaryTitle string     `json:"primaryTitle"`
    ShowTimes    []ShowTime `json:"showTimes"`
}

type ShowTime struct {
    PerformanceID int    `json:"performanceId"`
    DisplayDate   string `json:"displayDate"`
    DisplayTime   string `json:"displayTime"`
    Venue         string `json:"locationVenueName"`
    SeatingStatus string `json:"seatingStatus"`
	LocationVenueBrandCode string `json:"locationVenueBrandCode"`
}
