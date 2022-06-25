# RNKM65 Auth

## Stacks
- golang
- gRPC

## Getting Start
These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites
- golang 1.18 or [later](https://go.dev)
- docker
- makefile

### Installing
1. Clone the project from [RNKM65 Auth](https://github.com/isd-sgcu/rnkm65-auth)
2. Import project
3. Copy `config.example.yaml` in `config` and paste it in the same location then remove `.example` from its name.
4. Download dependencies by `go mod download`

### Testing
1. Run `go test  -v -coverpkg ./... -coverprofile coverage.out -covermode count ./...` or `make test`

### Running
1. Run `docker-compose up -d` or `make compose-up`
2. Run `go run ./src/.` or `make server`

### Compile proto file
1. Run `make proto`

## Chula SSO Mock
1. Make sure you follow the `Running` step
2. Go to `http://localhost:8080/html/login.html?service=https://google.com`
3. Login (you can fill up with any things)
4. The ticket will be in the query param

## Special Thanks
Special thanks to [saengowp](https://github.com/saengowp) for [Chula SSO Mock](https://github.com/saengowp/chulassomock)
