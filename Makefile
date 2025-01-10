build:
	go build -o sonic

build-linux:
	GOOS=linux GOARCH=amd64 go build -o sonic .