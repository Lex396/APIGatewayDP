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
	api.router.HandleFunc("/check", api.checkHandler).Methods(http.MethodPost, http.MethodOptions)
}

func (api *API) checkHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка на запрещенные слова
	stopWords := []string{"qwerty", "йцукен", "zxvbnm"}
	contentLower := strings.ToLower(request.Content)

	for _, word := range stopWords {
		if strings.Contains(contentLower, strings.ToLower(word)) {
			log.Printf("Найдено запрещенное слово: %s", word)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
