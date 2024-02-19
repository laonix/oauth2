build-img:
	docker build -t oauth2:1.0.0 .

run-img: build-img
	docker run --rm -p 3000:3000 oauth2:1.0.0

run:
	go run cmd/main.go

test:
	go test -v -race -count=1 ./...

e2e:
	go run demo/e2e/main.go

rate:
	go run demo/rate/main.go