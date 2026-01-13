package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

// Register регистрация пользователя
//
// POST /api/user/register.
// Регистрация производится по паре логин/пароль. Каждый логин должен быть уникальным.
// После успешной регистрации должна происходить автоматическая аутентификация пользователя.
// Для передачи аутентификационных данных используйте механизм HTTP-заголовок Authorization.
//
// Коды ответов:
// 200 — пользователь успешно зарегистрирован и аутентифицирован;
// 400 — неверный формат запроса;
// 409 — логин уже занят;
// 500 — внутренняя ошибка сервера.
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.standardError(w, "failed to read body", err, http.StatusInternalServerError)
		return
	}

	var data struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			a.customError(w, "json syntax error", se, http.StatusBadRequest)
		} else {
			a.standardError(w, "unexpected json error", err, http.StatusInternalServerError)
		}
		return
	}

	u, err := a.userRepository.Register(data.Login, data.Password)
	if err != nil {
		if errors.Is(err, model.ErrUserAlreadyExists) {
			a.customError(w, fmt.Sprintf("user %s already exists", data.Login), err, http.StatusConflict)
		} else {
			a.standardError(w, "unexpected error during the registration", err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", u.Token))
}

// Login — аутентификация пользователя по паре логин/пароль
//
// POST /api/user/login.
// Для передачи аутентификационных данных используйте механизм HTTP-заголовок Authorization.
//
// Коды ответов:
// 200 — пользователь успешно аутентифицирован;
// 400 — неверный формат запроса;
// 401 — неверная пара логин/пароль;
// 500 — внутренняя ошибка сервера.
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.standardError(w, "failed to read body", err, http.StatusInternalServerError)
		return
	}

	var data struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			a.customError(w, "json syntax error", se, http.StatusBadRequest)
		} else {
			a.standardError(w, "unexpected json error", err, http.StatusInternalServerError)
		}
		return
	}

	u, err := a.userRepository.AuthByLogin(data.Login, data.Password)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			a.customError(w, fmt.Sprintf("user %s not found", data.Login), err, http.StatusUnauthorized)
		case errors.Is(err, model.ErrWrongPassword):
			a.customError(w, "wrong password", err, http.StatusUnauthorized)
		default:
			a.standardError(w, "unexpected error during the registration", err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer: %s", u.Token))
}

func (a *API) userFromContext(w http.ResponseWriter, r *http.Request) (*model.User, bool) {
	u, ok := r.Context().Value(userContextKey).(*model.User)
	if !ok || u == nil {
		a.logger.Error("unexpectedly empty user")
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	return u, true
}

// Balance - получение текущего баланса пользователя
//
// GET /api/user/balance
// Хендлер доступен только авторизованному пользователю.
// В ответе должны содержаться данные о текущей сумме баллов лояльности,
// а также сумме использованных за весь период регистрации баллов.
//
// Коды ответов:
// 200 — успешная обработка запроса;
// 401 — пользователь не авторизован;
// 500 — внутренняя ошибка сервера.
func (a *API) Balance(w http.ResponseWriter, r *http.Request) {
	u, ok := a.userFromContext(w, r)
	if !ok {
		return
	}

	type resData struct {
		Current   float64 `json:"current"`
		Withdrawn int     `json:"withdrawn"`
	}

	data, err := json.Marshal(resData{
		Current:   u.Balance,
		Withdrawn: u.Withdrawn,
	})
	if err != nil {
		a.standardError(w, "failed to marshal json", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Withdrawals - получение информации о выводе средств
// с накопительного счёта пользователем.
//
// GET /api/user/withdrawals
// Хендлер доступен только авторизованному пользователю.
// Факты выводов в выдаче должны быть отсортированы
// по времени вывода от самых новых к самым старым.
// Формат даты — RFC3339.
//
// Коды ответов:
// 200 — успешная обработка запроса;
// 204 — нет ни одного списания;
// 401 — пользователь не авторизован;
// 500 — внутренняя ошибка сервера.
func (a *API) Withdrawals(w http.ResponseWriter, r *http.Request) {
	u, ok := a.userFromContext(w, r)
	if !ok {
		return
	}
	ws, err := a.userRepository.Withdrawals(u.ID)
	if err != nil {
		a.standardError(w, "failed to fetch withdrawals", err, http.StatusInternalServerError)
		return
	}
	if len(ws) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type resDataItem struct {
		Order       string    `json:"order"`
		Sum         int       `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}

	var data []resDataItem

	for _, w := range ws {
		data = append(data, resDataItem{
			Order:       w.OrderID,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}

	b, err := json.Marshal(data)
	if err != nil {
		a.standardError(w, "failed to json marshal", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// AddWithdrawal - запрос на списание баллов
// с накопительного счёта в счёт оплаты нового заказа&
//
// POST /api/user/balance/withdraw
// Хендлер доступен только авторизованному пользователю.
// Номер заказа представляет собой гипотетический номер нового заказа пользователя,
// в счёт оплаты которого списываются баллы.
//
// Коды ответов:
// 200 — успешная обработка запроса;
// 401 — пользователь не авторизован;
// 402 — на счету недостаточно средств;
// 422 — неверный номер заказа;
// 500 — внутренняя ошибка сервера.
func (a *API) AddWithdrawal(w http.ResponseWriter, r *http.Request) {
	u, ok := a.userFromContext(w, r)
	if !ok {
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.standardError(w, "failed to read body", err, http.StatusInternalServerError)
		return
	}

	// todo - вынести здесь и в аналогичных местах - в отдельный тип
	var data struct {
		OrderID string `json:"order"`
		Sum     int    `json:"sum"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			a.customError(w, "json syntax error", se, http.StatusBadRequest)
		} else {
			a.standardError(w, "unexpected json error", err, http.StatusInternalServerError)
		}
		return
	}

	err = a.userRepository.AddWithdrawal(u.ID, data.OrderID, data.Sum)
	if err != nil {
		if errors.Is(err, model.ErrInsufficientFunds) {
			a.standardError(w, "insufficient funds", err, http.StatusPaymentRequired)
			return
		}
		a.standardError(w, "unexpected error", err, http.StatusInternalServerError)
		return
	}
}

/*
POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
*/
