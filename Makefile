
.PHONY: help
help:
	go run goi18n/*.go -h
	go run goi18n/*.go walk -h

.PHONY: http
http:
	go run goi18n/*.go walk --verbose --path=example/http

b:
	rm -rf example/http/z
	go build -o example/http/z ./goi18n/...
	./example/http/z --verbose walk --path=example/http
