package routes

import (
	"net/http"
	"time"

	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
)

type createReservationRequest struct {
	SeatIDs []int `json:"seat_ids"`
}

func RegisterReservationRoutes(r *gin.Engine) {
	r.POST("/showtimes/:id/reservations", AuthMiddleware(), createReservation)

	r.GET("/me/reservations", AuthMiddleware(), listUserReservations)

	r.PUT("/reservations/:id/cancel", AuthMiddleware(), cancelReservation)
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
	//preventing booking pastime shows
	var startTime time.Time

	err := db.DB.QueryRow(`
		SELECT start_time
		FROM showtimes
		WHERE id =$1
		`, showtimeID).Scan(&startTime)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
		return
	}

	if startTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot book past showtime"})
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

func listUserReservations(c *gin.Context) {

	// 1. Identity Check: "Who is asking?"
	// We trust the AuthMiddleware. It already checked the token and put the ID here.
	userID := c.GetInt("user_id")

	// 2. The "Reporter" Query: Gather data from 3 different places.
	rows, err := db.DB.Query(`
		SELECT 
			r.id, 		--reservation ticket #
			m.title,	--Movie Name(from movies table)
			s.start_time,--Time (from showtimes table)
			r.status	--"BOOKED" or "CANCELLED"
		FROM reservations r
		JOIN showtimes s ON s.id = r.showtime_id --Connect Ticket -> Time
		JOIN movies m ON m.id = s.movie_id		 --Connect Time -> Movie Name
		WHERE r.user_id = $1					 --Filter: Only showe this user's ticket
		ORDER BY r.created_at DESC				 --Sort : Show newest booking at the top
		`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// 3. The Empty List: Prepare a container for the results.
	var results []gin.H

	// 4. The Loop: Process the pile of receipts one by one.
	for rows.Next() {
		var id int
		var title string
		var startTime string
		var status string

		// Scan the 4 columns we asked for in the SELECT statement.
		rows.Scan(&id, &title, &startTime, &status)

		// Add to our list
		results = append(results, gin.H{
			"reservation_id": id,
			"movie_title":    title,
			"start_time":     startTime,
			"status":         status,
		})
	}

	// 5. Send the Report

	c.JSON(http.StatusOK, results)
}

func cancelReservation(c *gin.Context) {
	// 1. Identify the Target
	// Which reservation are we cancelling? (from URL)
	resID := c.Param("id")
	// Who is asking? (from Token)
	userID := c.GetInt("user_id")

	// 2. Gather Evidence (The Validation Query)
	// Before we do anything, we need to look up the reservation details.
	// We need to know:
	//   a. When is the movie? (showtimeTime) - Used for policy checks (e.g., can't cancel if movie started).
	//   b. Who owns it? (ownerID) - Security check.
	//   c. What is the current status? (status) - Logic check.
	var showtimeTime time.Time
	var ownerID int
	var status string

	err := db.DB.QueryRow(`
		SELECT s.start_time, r.user_id, r.status
		FROM reservations r
		JOIN showtimes s ON s.id = r.showtime_id
		WHERE r.id = $1
		`, resID).Scan(&showtimeTime, &ownerID, &status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reservastion not found"})
		return
	}

	//fixing cancel past showtime
	if showtimeTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot cancel past showtimes",
		})
		return
	}

	// 3. The Security Check ("Is this yours?")
	// If the ID on the ticket (ownerID) doesn't match the ID in the token (userID),
	// it means Hacker Bob is trying to cancel Alice's ticket. Block them.
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your reservation"})
		return
	}

	// 4. The Logic Check ("Is it valid?")
	// You can't cancel a ticket that is already cancelled or refunded.
	if status != "BOOKED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannnot cancel"})
		return
	}

	// 5. The Action (Soft Delete)
	// We do NOT use "DELETE FROM". We use "UPDATE".
	// Why? Because accountants need records! We want to keep the history that
	// "User X booked this, then cancelled it."
	_, err = db.DB.Exec(`
		UPDATE reservations
		SET status = 'CANCELLED'
		WHERE id = $1
		`, resID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reservation cancelled"})
}
