package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"APIGateway/aggregator/pkg/logger"
	"APIGateway/aggregator/pkg/storage"
)

const postsPerPage = 5

type API struct {
	DB     storage.StorageInterface
	Logger *logger.Logger
	router *mux.Router
}

// NewAPI создает экземпляр API.
func NewAPI(db storage.StorageInterface, log *logger.Logger) *API {
	api := &API{
		DB:     db,
		Logger: log,
		router: mux.NewRouter(),
	}
	api.endpoints()
	return api
}

func (api *API) Router() *mux.Router {

	return api.router
}

// endpoints регистрирует маршруты.
func (a *API) endpoints() {
	a.router.HandleFunc("/news", a.postsHandler).Methods(http.MethodGet, http.MethodOptions)
	a.router.HandleFunc("/news/latest", a.newsLatestHandler).Methods(http.MethodGet, http.MethodOptions)
	a.router.HandleFunc("/news/search", a.newsDetailedHandler).Methods(http.MethodGet, http.MethodOptions)

}

// postsHandler Возвращает все публикации.
func (a *API) postsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	pageParam := r.URL.Query().Get("page")
	if pageParam == "" {
		pageParam = "1"
	}
	sParam := r.URL.Query().Get("s")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page <= 0 {
		http.Error(w, "invalid page", http.StatusBadRequest)
		return
	}

	limit := postsPerPage
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil || limit <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
	}

	posts, pagination, err := a.DB.GetPostsByTitle(sParam, limit, (page-1)*limit)
	if err != nil {
		log.Printf("DB error in postsHandler: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Posts      []storage.Post     `json:"posts"`
		Pagination storage.Pagination `json:"pagination"`
	}{
		Posts:      posts,
		Pagination: pagination,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// newsLatestHandler получение страницы с определенным номером
func (api *API) newsLatestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	pageParam := r.URL.Query().Get("page")
	if pageParam == "" {
		pageParam = "1"
	}
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		http.Error(w, "Неверный параметр страницы", http.StatusBadRequest)
		return
	}

	posts, pagination, err := api.DB.GetLastPosts(postsPerPage, page)
	if err != nil {
		http.Error(w, "Ошибка получения публикаций", http.StatusInternalServerError)
		return
	}

	response := struct {
		Posts      []storage.Post     `json:"posts"`
		Pagination storage.Pagination `json:"pagination"`
	}{
		Posts:      posts,
		Pagination: pagination,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// newsDetailedHandler получение публикации по id
func (api *API) newsDetailedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "Отсутствует параметр id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Неверный формат id", http.StatusBadRequest)
		return
	}

	post, err := api.DB.GetPostByID(id)
	if err != nil {
		if err.Error() == "post not found" {
			http.Error(w, "пост не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка при получении поста", http.StatusInternalServerError)
		}
		return
	}

	type RenderPost struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Content string `json:"content"`
		PubTime string `json:"pub_time"`
	}

	renderPost := RenderPost{
		Title:   post.Title,
		Link:    post.Link,
		Content: trimContent(stripHTML(post.Content), 1000),
		PubTime: formatDate(post.PubTime),
	}

	response := map[string]interface{}{
		"post": renderPost,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.Logger.Error("Ошибка отправки ответа:", err)
		http.Error(w, "Ошибка рендеринга JSON", http.StatusInternalServerError)
	}
}

// stripHTML удаляет HTML-теги из строки.
func stripHTML(input string) string {
	re := regexp.MustCompile("<.*?>")
	return strings.TrimSpace(re.ReplaceAllString(input, ""))
}

// trimContent обрезает текст до maxLen символов.
func trimContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}

// formatDate форматирует дату в нужный вид.
func formatDate(t time.Time) string {
	return t.Format("02.01.2006, 15:04:05")
}
