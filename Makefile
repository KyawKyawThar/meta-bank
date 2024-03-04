DB_URL=postgresql://root:secret@localhost:5432/hl-bank?sslmode=disable

postgres:
	docker run --name postgres16.1 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16.1-alpine
create_db:
	docker exec -it postgres16.1 createdb --username=root --owner=root hl-bank
drop_db:
	docker exec -it postgres16.1 dropdb hl-bank
new_migration:
	migrate create -ext sql -dir db/migrations -seq $(name)
migrate_up:
	migrate -path db/migrations -database $(DB_URL) -verbose up
migrate_down:
	migrate -path db/migrations -database $(DB_URL) -verbose down
test:
	go test -v -cover ./...
sqlc:
	sqlc generate


.PHONY:postgres create_db drop_db new_migration migrate_up migrate_down sqlc test