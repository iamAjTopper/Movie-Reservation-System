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

func updateMovie(c *gin.Context) {
	// 1. Capture the Target.
	// We expect the URL to look like "/admin/movies/5".
	// c.Param("id") grabs that "5" so we know WHICH movie to update.
	id := c.Param("id")

	// 2. Capture the New Data.
	// We reuse the 'createMovieRequest' struct because the data fields (Title, Genre) are the same.
	var req createMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 3. The Safety Check (Verification).
	// Before we try to update, let's make sure the movie actually exists.
	// "SELECT EXISTS" is a very fast SQL query that returns true/false.
	var exists bool
	err := db.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM movies WHERE id =$1)`,
		id,
	).Scan(&exists)

	// If the DB crashes (err) or the movie isn't there (!exists), stop.
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	// 4. The Update (Execution).
	// We use the SQL UPDATE command.
	// Notice the WHERE clause at the end—without it, you would overwrite EVERY movie in your database!
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

	// 5. Success
	c.JSON(http.StatusOK, gin.H{"message": "movie updated"})
}

func deleteMovie(c *gin.Context) {
	// 1. Identify the Target.
	// We grab the ID from the URL (e.g., "/movies/5").
	id := c.Param("id")

	// 2. The Attempt (Fire and Forget).
	// We run the DELETE command immediately.
	// Note: In SQL, if you try to delete ID 999 and it doesn't exist,
	// it is NOT an error! The database just says "Command successful, 0 items removed."
	result, err := db.DB.Exec(
		`DELETE FROM movies WHERE id = $1`,
		id,
	)

	// 3. Technical Error Check.
	// This only catches serious problems (e.g., DB is offline, SQL syntax error).
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. The "Ghost" Check (Crucial Step).
	// Since SQL doesn't error on missing IDs, we must manually ask:
	// "How many rows were actually touched by that last command?"
	rows, _ := result.RowsAffected()

	// 5. The Verdict.
	// If 0 rows were touched, it means the movie ID didn't exist in the first place.
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	//6. Success
	c.JSON(http.StatusOK, gin.H{"message": "movie deleted"})
}
