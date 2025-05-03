package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

// RateRequest структура для парсинга запроса изменения скорости
// @Description Запрос на изменение скорости пополнения токенов для клиента
type RateRequest struct {
	// @Description ID клиента
	// @Required
	ClientID string `json:"clientId"`
	// @Description Скорость пополнения токенов (в токенах в секунду)
	// @Required
	// @Minimum 0
	Rate int `json:"rate"`
}

// CapacityRequest структура для парсинга запроса изменения емкости
// @Description Запрос на изменение максимальной емкости корзины токенов для клиента
type CapacityRequest struct {
	// @Description ID клиента
	// @Required
	ClientID string `json:"clientId"`
	// @Description Максимальная емкость корзины токенов
	// @Required
	// @Minimum 1
	Capacity int `json:"capacity"`
}

// SetClientRate устанавливает скорость пополнения токенов для клиента
// @Summary Установка скорости пополнения токенов
// @Description Устанавливает скорость, с которой пополняется корзина токенов клиента
// @Tags клиенты
// @Accept json
// @Produce json
// @Param request body RateRequest true "Данные для установки скорости"
// @Success 200 {object} map[string]string "Успешный ответ"
// @Failure 400 {object} model.ErrorResponse "Ошибка в запросе"
// @Failure 405 {object} model.ErrorResponse "Метод не разрешен"
// @Router /client/rate [post]
func (h *Handler) SetClientRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req RateRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, ErrClientIDRequired)
		return
	}

	h.storage.SetClientRate(req.ClientID, req.Rate)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// SetClientCapacity устанавливает максимальную емкость корзины токенов для клиента
// @Summary Установка емкости корзины токенов
// @Description Устанавливает максимальное количество токенов, которое может накопить клиент
// @Tags клиенты
// @Accept json
// @Produce json
// @Param request body CapacityRequest true "Данные для установки емкости"
// @Success 200 {object} map[string]string "Успешный ответ"
// @Failure 400 {object} model.ErrorResponse "Ошибка в запросе"
// @Failure 405 {object} model.ErrorResponse "Метод не разрешен"
// @Router /client/capacity [post]
func (h *Handler) SetClientCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req CapacityRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, ErrClientIDRequired)
		return
	}

	if req.Capacity < 0 {
		newErrorResponse(w, r, http.StatusBadRequest, ErrCapacityRequired)
		return
	}

	h.storage.SetClientCapacity(req.ClientID, req.Capacity)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetClient получает информацию о клиенте
// @Summary Получение информации о клиенте
// @Description Возвращает информацию о клиенте по его ID
// @Tags клиенты
// @Produce json
// @Param clientId query string true "ID клиента"
// @Success 200 {object} model.Client "Информация о клиенте"
// @Failure 400 {object} model.ErrorResponse "Ошибка в запросе"
// @Failure 404 {object} model.ErrorResponse "Клиент не найден"
// @Failure 405 {object} model.ErrorResponse "Метод не разрешен"
// @Router /client [get]
func (h *Handler) GetClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, ErrClientIDRequired)
		return
	}

	client, ok := h.storage.GetClient(clientID)
	if !ok {
		newErrorResponse(w, r, http.StatusNotFound, ErrClientNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func parseJSONBody(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return ErrInvalidContentType
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return ErrReadRequestBody
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return ErrInvalidJSONFormat
	}

	return nil
}
