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
}

func createMovie(c *gin.Context) {
	var req createMovieRequest

	// 1. Parsing: Unpack the box.
	// We take the JSON sent by the admin and try to fit it into our struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 2. Validation: Check the rules.
	// A movie MUST have a title. If it's empty, reject it.
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	// 3. Execution: Write to the Ledger (Database).
	// We use Exec() because we are changing data, not asking for it.
	// standard SQL INSERT statement with $1, $2, etc. as placeholders.
	_, err := db.DB.Exec(
		`INSERT INTO movies (title, description, genre, poster_url)
		VALUES($1, $2, $3, $4)`,
		req.Title,
		req.Description,
		req.Genre,
		req.PosterURL,
	)

	// 4. Error Check: Did the database complain?
	// Maybe the database is down, or the title is too long for the column.
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5. Success: Confirm the action.
	// 201 Created is the standard HTTP code for "I built what you asked for."
	c.JSON(http.StatusCreated, gin.H{"message": "movie created"})
}

func listMovies(c *gin.Context) {
	// 1. The Request: "Give me everything."
	// We use Query() instead of QueryRow() because we expect MANY results,
	// not just one. This returns a "Rows" object (like a cursor).
	rows, err := db.DB.Query(
		`SELECT id,title,description,genre,poster_url FROM movies`,
	)

	// 2. The Check: Did the database crash?
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}

	// 3. The Cleanup: "Don't forget to hang up the phone."
	// Crucial! If you don't Close() the rows, the connection stays open forever.
	// Eventually, your database will run out of connections and crash.
	// 'defer' means: "Run this line at the very end of the function."
	defer rows.Close()

	// 4. The Container: A place to hold the results.
	// We create an empty list (slice) of generic JSON objects
	var movies []gin.H

	// 5. The Loop: "Next, please!"
	// rows.Next() advances the cursor to the next row in the database result.
	// It returns 'true' if there is data, and 'false' when the list is empty.
	for rows.Next() {
		// Create temporary variables to hold the data for THIS specific movie.
		var id int
		var title, description, genre, posterURL string

		// 6. The Copy: Move data from SQL to Go.
		// Just like before, the order must match the SELECT statement exactly.
		// id -> &id, title -> &title, etc.
		rows.Scan(&id, &title, &description, &genre, &posterURL)

		// 7. The Stack: Add it to our list.
		// We wrap the variables in a nice JSON object and add it to the 'movies' slice.

		movies = append(movies, gin.H{
			"id":          id,
			"title":       title,
			"description": description,
			"genre":       genre,
			"poster_url":  posterURL,
		})

	}
	// 8. The Delivery: Serve the full list.
	// We send the entire slice back to the user with a 200 OK status.
	c.JSON(http.StatusOK, movies)
}
