package routes

import (
	"movie-reservation/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createMovieRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
	PosterURL   string `json:"poster_url"`
}

func RegisterMovieRoutes(r *gin.Engine) {
	//PUBLIC
	r.GET("/movies", listMovies)

	//admin
	r.POST("/movies", AuthMiddleware(), AdminOnlyMiddleware(), createMovie)

	r.PUT("/movies/:id", AuthMiddleware(), AdminOnlyMiddleware(), updateMovie)

	r.DELETE("/movies/:id", AuthMiddleware(), AdminOnlyMiddleware(), deleteMovie)
}

func createMovie(c *gin.Context) {
	var req createMovieRequest

	// 1 Parsing
	// We are taking the JSON sent by the admin and try to fit it into our struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 2 Validation
	// A movie MUST have a title, If it's empty, reject it
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	// 3 Execution: Write to the Databse

	_, err := db.DB.Exec(
		`INSERT INTO movies (title, description, genre, poster_url)
		VALUES($1, $2, $3, $4)`,
		req.Title,
		req.Description,
		req.Genre,
		req.PosterURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5 Success

	c.JSON(http.StatusCreated, gin.H{"message": "movie created"})
}

func listMovies(c *gin.Context) {
	// 1. The Request

	rows, err := db.DB.Query(
		`SELECT id,title,description,genre,poster_url FROM movies`,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}

	defer rows.Close()

	// 4. The Container: We create an empty list (slice) of generic JSON objects
	var movies []gin.H

	// 5. The Loop: advaces to next datatbse
	// It returns 'true' if there is data, and 'false' when the list is empty
	for rows.Next() {
		// Create temporary variables to hold the data for THIS specific movie.
		var id int
		var title, description, genre, posterURL string

		// 6 The Copy: Move data from SQL to Go

		rows.Scan(&id, &title, &description, &genre, &posterURL)

		// 7 The Stack: Add it to our list
		// We wrap the variables in a nice JSON object and add it to the 'movies' slice

		movies = append(movies, gin.H{
			"id":          id,
			"title":       title,
			"description": description,
			"genre":       genre,
			"poster_url":  posterURL,
		})

	}
	c.JSON(http.StatusOK, movies)
}

func updateMovie(c *gin.Context) {
	// 1 Capture the id
	id := c.Param("id")

	// 2 Capture the New Data
	var req createMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 3 The Safety Check

	var exists bool
	err := db.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM movies WHERE id =$1)`,
		id,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	// 4 The Update

	_, err = db.DB.Exec(
		`UPDATE movies
		SET title = $1, description = $2, genre = $3, poster_url = $4
		WHERE id = $5`,
		req.Title,
		req.Description,
		req.Genre,
		req.PosterURL,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5 Success
	c.JSON(http.StatusOK, gin.H{"message": "movie updated"})
}

func deleteMovie(c *gin.Context) {
	// 1 Identify the id

	id := c.Param("id")

	// 2 The Attempt: We run the DELETE command

	result, err := db.DB.Exec(
		`DELETE FROM movies WHERE id = $1`,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4 The "Ghost" Check: How many rows were actually touched by that last command?

	rows, _ := result.RowsAffected()

	// 5  If 0 rows were touched, it means the movie ID didn't exist in the first place.

	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	//6 Success
	c.JSON(http.StatusOK, gin.H{"message": "movie deleted"})
}
