martd: *.go
	go build

static.go: index.html client.js
	go generate

