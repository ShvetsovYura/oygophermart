build: app

app:
	cd cmd/gophermart && go build -o gophermart *.go

t:
	go test ./...
