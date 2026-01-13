

curl -X POST -i \
'http://localhost:8080/api/user/register' \
-d '{
    "login": "mike",
    "password": "some"
}' \
| ijq


curl -X GET -i \
'http://localhost:8080/api/user/balance' \
-H 'Authorization: Bearer 87bf7c2e-ef2c-11f0-b29b-96f332c237f9' \
| ijq
