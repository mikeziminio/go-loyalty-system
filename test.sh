set -ex

go vet -vettool=./bin/statictest ./...
echo "Verified by linter"

./bin/gophermarttest \
    -test.v -test.run=^TestGophermart$ \
    -gophermart-binary-path=./bin/gophermart \
    -gophermart-host=localhost \
    -gophermart-port=8080 \
    -gophermart-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable" \
    -accrual-binary-path=cmd/accrual/accrual_darwin_amd64 \
    -accrual-host=localhost \
    -accrual-port=$(./bin/random unused-port) \
    -accrual-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"
