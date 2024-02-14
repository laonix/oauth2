build-img:
	docker build -t oauth:1.0.0 .

run-img: build-img
	docker run --rm -p 3000:3000 oauth:1.0.0

run:
	go run cmd/main.go

test:
	go test -v -race -count=1 ./...