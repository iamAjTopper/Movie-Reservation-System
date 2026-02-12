package routes

import (
	"net/http"

	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
)

type createReservationRequest struct {
	SeatIDs []int `json:"seat_ids"`
}

func RegisterReservationRoutes(r *gin.Engine) {
	r.POST("/showtimes/:id/reservations", AuthMiddleware(), createReservation)
}

func createReservation(c *gin.Context) {
	// 1. Get the Context
	// Who is buying? (userID from the token)
	// What movie? (showtimeID from the URL)
	showtimeID := c.Param("id")
	userID := c.GetInt("user_id")

	// 2. Get the Cart (The specific seats they want)
	var req createReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 3. Validation
	// You can't buy 0 tickets.
	if len(req.SeatIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no seats selected"})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
	}
	defer tx.Rollback()

	// 4. The Availability Check (The Flaw is Here)
	// The code loops through every seat the user wants (e.g., Seat 1, Seat 2).
	// For each seat, it asks the database: "Is this seat taken?"
	for _, seatID := range req.SeatIDs {
		var available bool

		err := tx.QueryRow(`
			SELECT NOT EXISTS (
				SELECT 1
				FROM reserved_seats rs
				JOIN reservations r
					ON r.id = rs.reservation_id
				WHERE rs.seat_id = $1
				AND r.showtime_id = $2
				AND r.status = 'BOOKED'
			)
		`, seatID, showtimeID).Scan(&available)

		if err != nil || !available {
			c.JSON(http.StatusBadRequest, gin.H{"error": "one or more seats unavaliable"})
			return
		}
	}

	// 5. Create the Receipt (Reservation Record)
	// If the loop finished, the code assumes all seats are free.
	// It creates a new Reservation ID (e.g., #505)
	var reservationID int
	err = tx.QueryRow(`
		INSERT INTO reservations (user_id, showtime_id, status)
		VALUES ($1, $2, 'BOOKED')
		RETURNING id`, userID, showtimeID).Scan(&reservationID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6. Assign the Seats
	// It loops through the seats again and permanently links them to Reservation #505.
	for _, seatID := range req.SeatIDs {
		_, err := tx.Exec(`
			INSERT INTO reserved_seats (reservation_id, seat_id)
			VALUES ($1, $2)`, reservationID, seatID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reserve seats"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"reservation_id": reservationID,
	})
}
