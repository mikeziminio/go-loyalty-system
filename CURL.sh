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

# получение текущего баланса счёта баллов лояльности пользователя
curl -X GET -i \
'http://localhost:8080/api/user/balance' \
-H 'Authorization: Bearer d97d7c2e-f2e6-11f0-9b2c-96f332c237f9' \
| ijq

# запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
curl -X POST -i \
'http://localhost:8080/api/user/balance/withdraw' \
-H 'Authorization: Bearer d97d7c2e-f2e6-11f0-9b2c-96f332c237f9' \
-d '{
	"order": "2377225624",
    "sum": 5
}' \
| ijq

# получение информации о выводе средств с накопительного счёта пользователем.
curl -X GET -i \
'http://localhost:8080/api/user/withdrawals' \
-H 'Authorization: Bearer d97d7c2e-f2e6-11f0-9b2c-96f332c237f9' \
| ijq

# загрузка пользователем номера заказа для расчёта;
curl -X POST -i \
'http://localhost:8080/api/user/orders' \
-H 'Authorization: Bearer d97d7c2e-f2e6-11f0-9b2c-96f332c237f9' \
-H 'Content-Type: text/plain'
-d '12345888' \
| ijq

# получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
curl -X GET -i \
'http://localhost:8080/api/user/orders' \
-H 'Authorization: Bearer f9222256-f3a3-11f0-a16a-96f332c237f9' \
| ijq

#
# Запросы во внешнюю систему расчета баллов
#

# получение расчета для заказа из внешней системы
curl -X GET -i \
'http://localhost:8088/api/orders/49927398716' \
| ijq

# регистрация нового совершённого заказа
curl -X POST -i \
'http://localhost:8088/api/orders' \
-H 'Content-Type: application/json' \
-d '{
    "order": "49927398716",
    "goods": [
        {
            "description": "Чайник Bork",
            "price": 7000
        }
    ]
}'

# регистрация информации о новой механике вознаграждения за товар
curl -X POST -i \
'http://localhost:8088/api/goods' \
-H 'Content-Type: application/json' \
-d '{
    "match": "Bork",
    "reward": 10,
    "reward_type": "%"
}'
