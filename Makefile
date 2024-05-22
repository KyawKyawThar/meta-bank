DB_URL=postgresql://root:secret@localhost:5432/meta-bank?sslmode=disable

DB_URL_Docker=postgresql://root:secret@postgres:5432/meta-bank?sslmode=disable

network:
	docker network create bank-network
postgres:
	docker run --name postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16.2-alpine
create_db:
	docker exec -it postgres createdb --username=root --owner=root meta-bank
drop_db:
	docker exec -it meta-bank dropdb meta-bank
new_migration:
	migrate create -ext sql -dir db/migrations -seq $(name)
migrate_up:
	migrate -path db/migrations -database $(DB_URL_Docker) -verbose up
migrate_down:
	migrate -path db/migrations -database $(DB_URL_Docker) -verbose down
migrate_up1:
	migrate -path db/migrations -database $(DB_URL_Docker) -verbose up 1
migrate_down1:
	migrate -path db/migrations -database $(DB_URL_Docker) -verbose down 1
test:
	go test -v -cover ./...
sqlc:
	sqlc generate
#go:
#	docker build -t meta-bank:latest .
#go_run:
#	docker run --rm --name meta-bank --network bank-network -p 8080:8080 -e DB_SOURCE=$(DB_URL_Docker) -e GIN_MODE=release meta-bank:latest
.PHONY:postgres create_db drop_db new_migration migrate_up migrate_down sqlc test migrate_up1 migrate_down1 network