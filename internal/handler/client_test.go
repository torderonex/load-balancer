package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	mockStorage "github.com/torderonex/load-balancer/internal/mocks"
	"github.com/torderonex/load-balancer/internal/storage"
)

func TestSetClientRate(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    RateRequest
		contentType    string
		expectedStatus int
		mockSetup      func(mockStorage *mockStorage.MockClientStorage)
	}{
		{
			name:   "Set Rate Successfully",
			method: http.MethodPost,
			requestBody: RateRequest{
				ClientID: "127.0.0.1",
				Rate:     50,
			},
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			mockSetup: func(mockStorage *mockStorage.MockClientStorage) {
				mockStorage.On("SetClientRate", "127.0.0.1", 50).Return()
			},
		},
		{
			name:   "Empty ClientID",
			method: http.MethodPost,
			requestBody: RateRequest{
				ClientID: "",
				Rate:     50,
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
		{
			name:           "Wrong Method",
			method:         http.MethodGet,
			requestBody:    RateRequest{},
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
		{
			name:   "Wrong Content Type",
			method: http.MethodPost,
			requestBody: RateRequest{
				ClientID: "127.0.0.1",
				Rate:     50,
			},
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientStorage := mockStorage.NewMockClientStorage(t)
			tt.mockSetup(mockClientStorage)

			storageImpl := &storage.Storage{ClientStorage: mockClientStorage}
			h := &Handler{storage: storageImpl}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/client/rate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			h.SetClientRate(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "success", response["status"])
			}
		})
	}
}

func TestSetClientCapacity(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    CapacityRequest
		contentType    string
		expectedStatus int
		mockSetup      func(mockStorage *mockStorage.MockClientStorage)
	}{
		{
			name:   "Set Capacity Successfully",
			method: http.MethodPost,
			requestBody: CapacityRequest{
				ClientID: "127.0.0.1",
				Capacity: 200,
			},
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			mockSetup: func(mockStorage *mockStorage.MockClientStorage) {
				mockStorage.On("SetClientCapacity", "127.0.0.1", 200).Return()
			},
		},
		{
			name:   "Empty ClientID",
			method: http.MethodPost,
			requestBody: CapacityRequest{
				ClientID: "",
				Capacity: 200,
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
		{
			name:   "Capacity Less Than 0",
			method: http.MethodPost,
			requestBody: CapacityRequest{
				ClientID: "127.0.0.1",
				Capacity: -1,
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
		{
			name:           "Wrong Method",
			method:         http.MethodGet,
			requestBody:    CapacityRequest{},
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
		{
			name:   "Wrong Content Type",
			method: http.MethodPost,
			requestBody: CapacityRequest{
				ClientID: "127.0.0.1",
				Capacity: 200,
			},
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(mockStorage *mockStorage.MockClientStorage) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientStorage := mockStorage.NewMockClientStorage(t)
			tt.mockSetup(mockClientStorage)

			storageImpl := &storage.Storage{ClientStorage: mockClientStorage}
			h := &Handler{storage: storageImpl}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/client/capacity", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			h.SetClientCapacity(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "success", response["status"])
			}
		})
	}
}

func TestParseJSONBody(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		requestBody   string
		expectedError error
	}{
		{
			name:          "Valid JSON",
			contentType:   "application/json",
			requestBody:   `{"clientId": "127.0.0.1", "action": "ban"}`,
			expectedError: nil,
		},
		{
			name:          "Invalid Content Type",
			contentType:   "text/plain",
			requestBody:   `{"clientId": "127.0.0.1", "action": "ban"}`,
			expectedError: ErrInvalidContentType,
		},
		{
			name:          "Invalid JSON Format",
			contentType:   "application/json",
			requestBody:   `{"clientId": "127.0.0.1", "action": }`,
			expectedError: ErrInvalidJSONFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			var data map[string]interface{}
			err := parseJSONBody(req, &data)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "127.0.0.1", data["clientId"])
				assert.Equal(t, "ban", data["action"])
			}
		})
	}
}
