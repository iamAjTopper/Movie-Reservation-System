package routes

import (
	"net/http"

	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
)

type SeatAvailability struct {
	ID        int  `json:"id"`
	Row       int  `json:"row"`
	Number    int  `json:"number"`
	Available bool `json:"available"`
}

func RegisterSeatRoutes(r *gin.Engine) {
	r.GET("/showtimes/:id/seats", listSeatsForShowtime)
}

func listSeatsForShowtime(c *gin.Context) {
	showtimeID := c.Param("id")

	rows, err := db.DB.Query(`
        SELECT 
            s.id,
            s.row_label,
            s.seat_number,
            NOT EXISTS (
                SELECT 1
                FROM reserved_seats rs
                JOIN reservations r ON r.id = rs.reservation_id
                WHERE rs.seat_id = s.id
                AND r.showtime_id = $1
                AND r.status = 'BOOKED'
            ) AS available
        FROM seats s
        ORDER BY s.row_label, s.seat_number
    `, showtimeID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var seats []gin.H

	for rows.Next() {
		var id int
		var row string
		var number int
		var available bool

		rows.Scan(&id, &row, &number, &available)

		seats = append(seats, gin.H{
			"id":        id,
			"row":       row,
			"number":    number,
			"available": available,
		})
	}
	c.JSON(http.StatusOK, seats)
}
