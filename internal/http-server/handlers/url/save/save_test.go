package save

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func TestSaveHandler(t *testing.T) {
	log := slogdiscard.New()

	tests := []struct {
		name           string
		request        Request
		mockSetup      func(*mocks.URLSaverMock)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success with custom alias",
			request: Request{
				URL:   "https://google.com",
				Alias: "test_alias",
			},
			mockSetup: func(m *mocks.URLSaverMock) {
				m.SetSaveURLSuccess(1)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Success with auto-generated alias",
			request: Request{
				URL:   "https://google.com",
				Alias: "",
			},
			mockSetup: func(m *mocks.URLSaverMock) {
				m.SetSaveURLSuccess(1)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Alias already exists (user provided)",
			request: Request{
				URL:   "https://google.com",
				Alias: "existing_alias",
			},
			mockSetup: func(m *mocks.URLSaverMock) {
				m.SetSaveURLExistsError()
			},
			expectedStatus: http.StatusOK,
			expectedError:  "alias already exists",
		},
		{
			name: "Alias collision with auto-generated (retry success)",
			request: Request{
				URL:   "https://google.com",
				Alias: "",
			},
			mockSetup: func(m *mocks.URLSaverMock) {
				callCount := 0
				m.SaveURLFunc = func(urlToSave, alias string) (int64, error) {
					callCount++
					if callCount == 1 {
						return 0, storage.ErrURLExists // First call fails
					}
					return 1, nil // Second call succeeds
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid URL",
			request: Request{
				URL:   "not-a-url",
				Alias: "test_alias",
			},
			mockSetup:      func(m *mocks.URLSaverMock) {},
			expectedStatus: http.StatusOK,
			expectedError:  "field URL is not a valid URL",
		},
		{
			name: "Empty URL",
			request: Request{
				URL:   "",
				Alias: "test_alias",
			},
			mockSetup:      func(m *mocks.URLSaverMock) {},
			expectedStatus: http.StatusOK,
			expectedError:  "field URL is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockURLSaver := mocks.NewURLSaverMock(t)
			tt.mockSetup(mockURLSaver)

			// Create handler
			handler := New(log, mockURLSaver)

			// Prepare request
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response body
			var response Response
			body := rr.Body.Bytes()
			if len(body) == 0 {
				t.Fatalf("empty response body")
			}
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v, body: %s", err, string(body))
			}

			if tt.expectedError != "" {
				if response.Status != "Error" || response.Error != tt.expectedError {
					t.Errorf("expected error %q, got status=%q error=%q", tt.expectedError, response.Status, response.Error)
				}
			} else {
				if response.Status != "OK" {
					t.Errorf("expected status OK, got %q", response.Status)
				}
				if tt.request.Alias == "" {
					// For auto-generated alias, check it's generated
					if response.Alias == "" || len(response.Alias) != aliasLenght {
						t.Errorf("expected auto-generated alias of length %d, got %q", aliasLenght, response.Alias)
					}
				} else {
					// For custom alias, check it matches
					if response.Alias != tt.request.Alias {
						t.Errorf("expected alias %q, got %q", tt.request.Alias, response.Alias)
					}
				}
			}
		})
	}
}
