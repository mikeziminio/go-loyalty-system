

POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.

# регистрация пользователя
curl -X POST -i \
'http://localhost:8080/api/user/register' \
-d '{
    "login": "mike",
    "password": "some"
}' \
| ijq

# аутентификация пользователя
curl -X POST -i \
'http://localhost:8080/api/user/login' \
-d '{
    "login": "mike",
    "password": "some"
}' \
| ijq


curl -X GET -i \
'http://localhost:8080/api/user/balance' \
-H 'Authorization: Bearer d97d7c2e-f2e6-11f0-9b2c-96f332c237f9' \
| ijq
