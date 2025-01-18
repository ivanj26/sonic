build:
	go build -o sonic

build-linux:
	GOOS=linux GOARCH=amd64 go build -o sonic-linux-amd64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o sonic-windows-amd64.exe .