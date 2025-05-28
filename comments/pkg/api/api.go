// api/api.go
package api

import (
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/comments/pkg/storage"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type API struct {
	router *mux.Router
	db     storage.Interface
	log    *logger.Logger
}

func New(db storage.Interface, log *logger.Logger) *API {
	api := API{
		router: mux.NewRouter(),
		db:     db,
		log:    log, // <--- инициализация
	}
	api.endpoints()
	return &api
}

func (api *API) Router() *mux.Router {
	return api.router
}

func (api *API) endpoints() {
	api.router.HandleFunc("/comments", api.commentsHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/comments", api.addCommentHandler).Methods(http.MethodPost, http.MethodOptions)
}

func (api *API) commentsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	newsIDStr := r.URL.Query().Get("news_id")
	newsID, err := strconv.Atoi(newsIDStr)
	if err != nil {
		api.log.ErrorWithRequestID(requestID.(string), "[commentsHandler] invalid news_id parameter:", newsIDStr, err)
		http.Error(w, "Invalid news_id parameter", http.StatusBadRequest)
		return
	}

	comments, err := api.db.AllComments(newsID)
	if err != nil {
		api.log.ErrorWithRequestID(requestID.(string), "[commentsHandler] failed to get comments for newsID=", newsID, "error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(comments); err != nil {
		api.log.ErrorWithRequestID(requestID.(string), "[commentsHandler] failed to encode comments:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	api.log.InfoWithRequestID(requestID.(string), "[commentsHandler] served comments for newsID=", newsID)
}

func (api *API) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var c storage.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if c.Content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	if c.NewsID == 0 && c.ParentID == nil {
		http.Error(w, "Either news_id or parent_id must be specified", http.StatusBadRequest)
		return
	}

	if err := api.db.AddComment(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
