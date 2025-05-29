// pkg/api/api.go
package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type API struct {
	router *mux.Router
}

func New() *API {
	api := API{
		router: mux.NewRouter(),
	}
	api.endpoints()
	return &api
}

func (api *API) Router() *mux.Router {
	return api.router
}

func (api *API) endpoints() {
	api.router.HandleFunc("/censor", api.handleCensor).Methods(http.MethodPost, http.MethodOptions)
}

func (api *API) handleCensor(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	log.Printf("Filtering text: %s", request.Text)

	stopWords := []string{"qwerty", "йцукен", "zxvbnm"}
	filtered := request.Text

	for _, word := range stopWords {
		filtered = strings.ReplaceAll(filtered, word, "***")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8") // <- добавь charset=utf-8
	resp := map[string]string{
		"text": filtered,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
