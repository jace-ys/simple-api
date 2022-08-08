package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/jace-ys/simple-api/domain"
)

type MCUHandler struct {
	movies domain.MoviesService
}

func NewMCUHandler(movies domain.MoviesService) *MCUHandler {
	return &MCUHandler{
		movies: movies,
	}
}

func (h *MCUHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/movies", h.GetMovies).Methods(http.MethodGet)
	r.HandleFunc("/movies/{id}", h.GetMovie)
}

func (h *MCUHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.movies.GetMovies(r.Context())
	if err != nil {
		log.Printf("GetMovies request error: %s\n", err)
		switch {
		default:
			respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, movies)
}

func (h *MCUHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		respondError(w, http.StatusBadRequest, "Missing ID for movie")
		return
	}

	movieID, err := strconv.Atoi(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movie ID")
		return
	}

	movie, err := h.movies.GetMovie(r.Context(), movieID)
	if err != nil {
		log.Printf("GetMovie request error: %s [id = %d]\n", err, movieID)
		switch {
		case errors.Is(err, domain.ErrMovieNotFound):
			respondError(w, http.StatusNotFound, "Movie not found")
		default:
			respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, movie)
}

func respondError(w http.ResponseWriter, statusCode int, errMsg string) {
	respondJSON(w, statusCode, map[string]interface{}{
		"error": map[string]interface{}{
			"status":  statusCode,
			"message": errMsg,
		},
	})
}

func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling payload: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(data))
}
