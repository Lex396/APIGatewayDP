package rss

import (
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/aggregator/pkg/storage"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mockHTTPClient(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

func TestParseRSS(t *testing.T) {
	logInstance, err := logger.NewLogger("test.log")
	if err != nil {
		t.Errorf("ошибка создания логгера: %v", err)
	}

	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectError   bool
		expectedPosts int
	}{
		{
			name: "Успешный парсинг RSS",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				rssData := `<?xml version="1.0"?>
				<rss>
					<channel>
						<item>
							<title>Test News</title>
							<description>Test Content</description>
							<pubDate>Mon, 2 Jan 2006 15:04:05 -0700</pubDate>
							<link>http://example.com</link>
						</item>
					</channel>
				</rss>`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(rssData))
			},
			expectError:   false,
			expectedPosts: 1,
		},
		{
			name: "Ошибка HTTP-запроса",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			expectError:   true,
			expectedPosts: 0,
		},
		{
			name: "Ошибка парсинга XML",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<rss><channel><item><title>Bad XML"))
			},
			expectError:   true,
			expectedPosts: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server := httptest.NewServer(test.serverHandler)
			defer server.Close()

			posts, err := parseRSS(server.URL, logInstance)

			if test.expectError {
				if err == nil {
					t.Errorf("Ожидалась ошибка, но её нет")
				}
			} else {
				if err != nil {
					t.Errorf("Не ожидалось ошибки, но получили: %v", err)
				}
				if len(posts) != test.expectedPosts {
					t.Errorf("Ожидалось %d постов, но получили %d", test.expectedPosts, len(posts))
				}
			}
		})
	}
}

func TestParsePubDate(t *testing.T) {
	logInstance, err := logger.NewLogger("test.log")
	if err != nil {
		t.Errorf("ошибка создания логгера: %v", err)
	}

	tests := []struct {
		dateStr      string
		shouldBeNow  bool
		description  string
		expectedTime time.Time
	}{
		{
			dateStr:      "Mon, 2 Jan 2006 15:04:05 -0700",
			shouldBeNow:  false,
			description:  "Корректный формат даты",
			expectedTime: time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("-0700", -7*60*60)),
		},
		{
			dateStr:      "invalid-date",
			shouldBeNow:  true,
			description:  "Некорректный формат даты",
			expectedTime: time.Now(),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			parsedTime := parsePubDate(test.dateStr, logInstance)

			if test.shouldBeNow {

				if time.Since(parsedTime) > time.Minute {
					t.Errorf("Ожидалось текущее время, но получено: %v", parsedTime)
				}
			} else {

				if !parsedTime.Equal(test.expectedTime) {
					t.Errorf("Ожидалось: %v, получено: %v", test.expectedTime, parsedTime)
				}
			}
		})
	}
}

func TestStartPolling(t *testing.T) {
	mockServer := mockHTTPClient(`
		<rss>
			<channel>
				<item>
					<title>Test News</title>
					<description>Test Description</description>
					<pubDate>Mon, 2 Jan 2006 15:04:05 -0700</pubDate>
					<link>http://example.com</link>
				</item>
			</channel>
		</rss>`, http.StatusOK)
	defer mockServer.Close()

	postChan := make(chan storage.Post, 1)
	errChan := make(chan error, 1)
	logInstance, err := logger.NewLogger("test.log")
	if err != nil {
		t.Errorf("ошибка создания логгера: %v", err)
	}

	go StartPolling([]string{mockServer.URL}, 1, postChan, errChan, logInstance)

	select {
	case post := <-postChan:
		if post.Title != "Test News" {
			t.Errorf("Ожидался заголовок 'Test News', получен: %s", post.Title)
		}
	case err := <-errChan:
		t.Fatalf("Получена ошибка: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Тайм-аут ожидания")
	}
}
