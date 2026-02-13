package routes

import (
	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
)

func AdminReportsRoutes(r *gin.Engine) {
	r.GET("/admin/report", AuthMiddleware(), AdminOnlyMiddleware(), adminReport)
}

func adminReport(c *gin.Context) {
	// 1 Run the "Big Picture" Query
	rows, err := db.DB.Query(`
		SELECT
			m.title,
			s.start_time,
			COUNT(rs.seat_id) AS seats_sold, -- 1 Count the tickets
			COUNT(rs.seat_id) * s.price AS revenue -- 2 Calculate the money
		FROM showtimes s
		JOIN movies m on m.id = s.movie_id
		--3 Using the left join trick
		LEFT JOIN reservations r
			ON r.showtime_id = s.id
			AND r.status = 'BOOKED'
		LEFT JOIN reserved_seats rs
			ON rs.reservation_id = r.id
		GROUP BY m.title, s.start_time, s.price
		ORDER BY s.start_time
		`)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var report []gin.H

	// 2 Processing the Results
	for rows.Next() {
		var title string
		var startTime string
		var seatsSold int
		var revenue float64

		// 3 Scanning the Data
		rows.Scan(&title, &startTime, &seatsSold, &revenue)

		report = append(report, gin.H{
			"movie_title": title,
			"start_time":  startTime,
			"seats_sold":  seatsSold,
			"revenue":     revenue,
		})

	}
	c.JSON(200, report)
}
