package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

/*
POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
*/

// Register регистрация пользователя
//
// Хендлер: POST /api/user/register.
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

	w.Header().Set("Authorization", fmt.Sprintf("Bearer: %s", u.Token))
}

// Login — аутентификация пользователя по паре логин/пароль
//
// Хендлер: POST /api/user/login.
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

// Balance - получение текущего баланса пользователя
//
// Хендлер: GET /api/user/balance
// Хендлер доступен только авторизованному пользователю.
// В ответе должны содержаться данные о текущей сумме баллов лояльности,
// а также сумме использованных за весь период регистрации баллов.
//
// Коды ответов:
// 200 — успешная обработка запроса;
// 401 — пользователь не авторизован;
// 500 — внутренняя ошибка сервера.
func (a *API) Balance(w http.ResponseWriter, r *http.Request) {

}
