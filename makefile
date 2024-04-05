postgres:
	docker run --network bank-network --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

createdb:
	docker exec -it postgres16 createdb --user=root --owner=root simple_bank

dropdb:
	docker exec -it postgres16 dropdb simple_bank

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
sqlc:
	sqlc generate

server:
	go run main.go

evans:
	evans --path proto --proto service_simple_bank.proto --host localhost --port 9090 repl

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/ashokmouli/simplebank/db/sqlc Store

proto:
	rm -rf proto/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
		proto/*.proto

tests: 
	go test -v -cover ./...
	
.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc server mock proto tests evans

