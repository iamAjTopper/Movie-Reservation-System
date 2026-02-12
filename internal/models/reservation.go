package models

import "time"

type Reservation struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ShowtimeID int       `json:"showtime_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type ReservedSeat struct {
	ReservationID int `json:"reservation_id"`
	SeatID        int `json:"seat_id"`
}
