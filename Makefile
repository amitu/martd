.PHONY: deps clean ping
msg=hello

./bin/martd: src/martd/static.go src/martd/*.go deps
	$(GOPATH)/bin/gb build all

src/martd/static.go: src/martd/index.html src/martd/client.js
	cd src/martd && go generate

clean:
	rm bin/martd

deps: $(GOPATH)/bin/esc $(GOPATH)/bin/gb

$(GOPATH)/bin/esc:
	go get github.com/mjibson/esc

$(GOPATH)/bin/gb:
	go get github.com/constabulary/gb/...

ping:
	curl -d "`date`: ${msg}" "http://localhost:54321/pub?channel=c1"
