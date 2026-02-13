package routes

import (
	"net/http"

	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"

	"time"
)

type createShowtimeRequest struct {
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Price     float64 `json:"price"`
}

func RegisterShowTimeRoutes(r *gin.Engine) {
	//public
	r.GET("/movies/:id/showtimes", listShowTimes)

	//admin
	r.POST("/movies/:id/showtimes", AuthMiddleware(), AdminOnlyMiddleware(), createShowtime)
}

func createShowtime(c *gin.Context) {
	// 1 Context:Which movie are we scheduling?

	movieID := c.Param("id")

	// 2 Data: When is it playing?
	var req createShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	//parse time

	startTime, err := time.Parse("2006-01-02T15:04:05", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format"})
		return
	}

	endTime, err := time.Parse("2006-01-02T15:04:05", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
		return
	}

	//end must be after the start
	if !endTime.After(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	//cannot scheule in the past
	if startTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time cannot be in the past"})
		return
	}

	//overlap proetcrtion
	var conflict bool
	err = db.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM showtimes
			WHERE movie_id = $1
			AND (
				(start_time < $3 and end_time > $2)
			)
		)`,
		movieID,
		startTime,
		endTime,
	).Scan(&conflict)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "validation failed"})
		return
	}

	if conflict {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "showtime overlaps with existing schedule",
		})
		return
	}
	// 3 Execution: Creating the link.
	// We insert a row that connects the movie_id (URL) with the times (Body)
	_, err = db.DB.Exec(
		`INSERT INTO showtimes (movie_id, start_time, end_time, price)
		VALUES ($1, $2, $3, $4)`,
		movieID,
		startTime,
		endTime,
		req.Price,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5 Success
	c.JSON(http.StatusCreated, gin.H{"message": "showtime created"})
}

func listShowTimes(c *gin.Context) {
	movieID := c.Param("id")

	rows, err := db.DB.Query(
		`SELECT id, start_time, end_time, price
		FROM showtimes
		WHERE movie_id = $1
		ORDER BY start_time`,
		movieID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch showtimes"})
		return
	}
	defer rows.Close()

	var showtimes []gin.H

	for rows.Next() {
		var id int
		var start, end string
		var price float64

		rows.Scan(&id, &start, &end, &price)

		showtimes = append(showtimes, gin.H{
			"id":         id,
			"start_time": start,
			"end_time":   end,
			"price":      price,
		})
	}
	c.JSON(http.StatusOK, showtimes)

}
