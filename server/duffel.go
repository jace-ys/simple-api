package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/jace-ys/simple-api/domain"
)

type DuffelFlightsHandler struct {
	airlineA domain.FlightsService
	airlineB domain.FlightsService
}

func NewDuffelFlightsHandler(airlineA, airlineB domain.FlightsService) *DuffelFlightsHandler {
	return &DuffelFlightsHandler{
		airlineA: airlineA,
		airlineB: airlineB,
	}
}

func (h *DuffelFlightsHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/flights/search", h.SearchFlights).Methods(http.MethodPost)
}

type SearchFlightsRequest struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
}

func (h *DuffelFlightsHandler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	body := &SearchFlightsRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	departureDate, err := time.Parse("2006-01-02", body.DepartureDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid departure date, must be of format YYYY-MM-DD")
		return
	}

	switch {
	case len(body.Origin) > 3:
		respondError(w, http.StatusBadRequest, "Invalid airport code for origin")
		return
	case len(body.Destination) > 3:
		respondError(w, http.StatusBadRequest, "Invalid airport code for destination")
		return
	}

	flights := domain.DuffelFlights{}

	flightsA, err := h.airlineA.GetFlights(r.Context(), body.Origin, body.Destination, departureDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("GetFlights request error: %s [airline = A]\n", err)
	} else {
		flights = append(flights, flightsA...)
	}

	flightsB, err := h.airlineB.GetFlights(r.Context(), body.Origin, body.Destination, departureDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("GetFlights request error: %s [airline = B]\n", err)
	} else {
		flights = append(flights, flightsB...)
	}

	if len(flights) == 0 {
		respondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("order")

	switch sortBy {
	case "price":
		flights = flights.SortByPrice(domain.SortOrder(sortOrder))
	case "duration":
		flights = flights.SortByDuration(domain.SortOrder(sortOrder))
	}

	respondJSON(w, http.StatusOK, flights)
}
