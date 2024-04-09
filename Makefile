DB_URL=postgresql://root:secret@localhost:5432/meta-bank?sslmode=disable

postgres:
	docker run --name meta-bank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16.2-alpine
create_db:
	docker exec -it meta-bank createdb --username=root --owner=root meta-bank
drop_db:
	docker exec -it meta-bank dropdb meta-bank
new_migration:
	migrate create -ext sql -dir db/migrations -seq $(name)
migrate_up:
	migrate -path db/migrations -database $(DB_URL) -verbose up
migrate_down:
	migrate -path db/migrations -database $(DB_URL) -verbose down
migrate_up1:
	migrate -path db/migrations -database $(DB_URL) -verbose up 1
migrate_down1:
	migrate -path db/migrations -database $(DB_URL) -verbose down 1
test:
	go test -v -cover ./...
sqlc:
	sqlc generate
redis:
	docker run --name redis -p 6379:6379 -d redis:7.2-alpine


.PHONY:postgres create_db drop_db new_migration migrate_up migrate_down sqlc test migrate_up1 migrate_down1