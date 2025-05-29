package api

import (
	"APIGateway/gateway/config"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type API struct {
	router      *mux.Router
	cfg         *config.Config
	newsURL     string
	censorURL   string
	commentsURL string
	client      *http.Client
}

func New(cfg *config.Config, newsURL, censorURL, commentsURL string) *API {
	api := &API{
		router:      mux.NewRouter(),
		cfg:         cfg,
		newsURL:     "http://localhost" + newsURL,
		censorURL:   "http://localhost" + censorURL,
		commentsURL: "http://localhost" + commentsURL,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
	api.initRoutes()
	return api
}

func (a *API) Router() *mux.Router {
	return a.router
}

func (a *API) initRoutes() {
	a.router.HandleFunc("/news", a.handleGetNews).Methods("GET")
	a.router.HandleFunc("/news/{id:[0-9]+}", a.handleGetNewsByID).Methods("GET")
	a.router.HandleFunc("/news/{id:[0-9]+}/comments", a.handlePostComment).Methods("POST")
}

func (a *API) handleGetNews(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", a.newsURL+"/news", nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.URL.RawQuery = r.URL.RawQuery
	copyHeader(r.Header, req.Header)

	resp, err := a.client.Do(req)
	if err != nil {
		http.Error(w, "failed to fetch news", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (a *API) handleGetNewsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	newsID := vars["id"]

	var newsData map[string]interface{}
	var commentsData []map[string]interface{}
	var wg sync.WaitGroup
	var newsErr, commentsErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		url := a.newsURL + "/news/" + newsID
		newsResp, err := a.client.Get(url)
		if err != nil {
			newsErr = err
			return
		}
		defer newsResp.Body.Close()
		if newsResp.StatusCode != http.StatusOK {
			newsErr = err
			return
		}
		if err := json.NewDecoder(newsResp.Body).Decode(&newsData); err != nil {
			newsErr = err
		}
	}()

	go func() {
		defer wg.Done()
		url := a.commentsURL + "/comments?news_id=" + newsID
		commentsResp, err := a.client.Get(url)
		if err != nil {
			commentsErr = err
			return
		}
		defer commentsResp.Body.Close()
		if commentsResp.StatusCode != http.StatusOK {
			commentsErr = err
			return
		}
		if err := json.NewDecoder(commentsResp.Body).Decode(&commentsData); err != nil {
			commentsErr = err
		}
	}()

	wg.Wait()

	if newsErr != nil {
		http.Error(w, "failed to get news", http.StatusBadGateway)
		return
	}
	if commentsErr != nil {
		http.Error(w, "failed to get comments", http.StatusBadGateway)
		return
	}

	result := map[string]interface{}{
		"news":     newsData,
		"comments": commentsData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (a *API) handlePostComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	newsID := vars["id"]
	log.Println("POST comment to news ID:", newsID)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var comment map[string]string
	if err := json.Unmarshal(body, &comment); err != nil {
		log.Println("Error unmarshalling JSON:", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	text := comment["text"]
	log.Println("Original comment text:", text)

	filteredText, err := a.filterText(text)
	if err != nil {
		log.Println("Error filtering text:", err)
		http.Error(w, "censorship failed", http.StatusBadGateway)
		return
	}
	comment["text"] = filteredText
	log.Println("Filtered comment text:", filteredText)

	jsonBody, _ := json.Marshal(comment)
	req, err := http.NewRequest("POST", a.commentsURL+"/comments?news_id="+newsID, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Error creating comment request:", err)
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	copyHeader(r.Header, req.Header)

	resp, err := a.client.Do(req)
	if err != nil {
		log.Println("Error sending comment:", err)
		http.Error(w, "failed to send comment", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	log.Println("Comment service response status:", resp.StatusCode)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (a *API) filterText(text string) (string, error) {
	log.Println("Filtering text:", text)

	requestBody := map[string]string{"text": text}
	jsonBody, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", a.censorURL+"/censor", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Error creating censor request:", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Println("Error sending censor request:", err)
		return "", err
	}
	defer resp.Body.Close()

	var respData map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		log.Println("Error decoding censor response:", err)
		return "", err
	}

	log.Println("Filtered text from censor:", respData["text"])
	return respData["text"], nil
}

func copyHeader(src http.Header, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
