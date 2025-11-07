APP_NAME=orders-api
IMAGE_NAME=orders-api:latest


setup:
	go mod tidy

run:
	go run ./cmd/api

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm -p 8080:8080 --name $(APP_NAME) $(IMAGE_NAME)

clean:
	rm -rf .reports
