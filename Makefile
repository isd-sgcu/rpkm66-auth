proto:
	protoc --proto_path=proto --go_out=. --go-grpc_out=require_unimplemented_servers=false:. user.proto
	protoc --proto_path=proto --go_out=. --go-grpc_out=require_unimplemented_servers=false:. auth.proto


test:
	go vet ./...
	go test  -v -coverpkg ./internal/... -coverprofile coverage.out -covermode count ./internal/...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

server:
	go run ./cmd/.

compose-up:
	docker-compose up -d

compose-down:
	docker-compose down

seed:
	go run ./cmd/. seed