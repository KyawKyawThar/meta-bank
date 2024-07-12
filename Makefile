DB_URL=postgresql://root:secret@localhost:5432/meta-bank?sslmode=disable

DB_URL_Docker=postgresql://root:secret@postgres:5432/meta-bank?sslmode=disable

DB_URL_RDS=postgresql://root:w0QUkLGYSPv1y6dkmLos@meta-bank.crsqiu0w479a.us-east-1.rds.amazonaws.com:5432/meta_bank

network:
	docker network create bank-network
postgres:
	docker run --name postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16.2-alpine
create_db:
	docker exec -it postgres createdb --username=root --owner=root meta-bank
drop_db:
	docker exec -it meta-bank dropdb meta-bank
rds_migration:
	migrate -path db/migrations -database $(DB_URL_RDS) -verbose up
new_migration:
	migrate create -ext sql -dir db/migrations -seq $(name)
migrate_up:
	migrate -path db/migrations -database $(DB_URL) -verbose up $(sequence)
migrate_down:
	migrate -path db/migrations -database $(DB_URL) -verbose down $(sequence)
test:
	go test -v -cover ./...
sqlc:
	sqlc generate
go:
	docker build -t meta-bank:latest .
go_run:
	docker run --rm --name meta-bank --network bank-network -p 8080:8080 -e DB_SOURCE=$(DB_URL_Docker) -e GIN_MODE=release meta-bank:latest

db_docs:
	dbdocs build doc/db.dbml --password secret
db_schema:
	dbml2sql doc/db.dbml -o doc/schema.sql
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/HL/meta-bank/db/sqlc Store
mock_source:
	mockgen -source=$(filePath)/db/sqlc/store.go -package mockdb -destination db/mock/store.go
#sample file path >>>> filePath=/Users/machineName/go/src/meta-bank this is my project file destination so run
# make mock_source filePath=/Users/machineName/go/src/meta-bank
# if you have trouble with this you can also use Reflect mode generates by mockgen running make mock
server:
	go run main.go


.PHONY:postgres create_db drop_db new_migration migrate_up migrate_down sqlc test network db_docs db_schema mock_source mock go go_run server

