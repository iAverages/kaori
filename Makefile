BINARY_NAME = kaori
DOCKER_IMAGE_NAME = ctr.avrg.dev/$(BINARY_NAME)
HASH = $(shell git rev-parse --short HEAD)

build:
	go build -o $(BINARY_NAME) -v ./cmd

docker:
	docker build -t $(DOCKER_IMAGE_NAME):$(HASH) .
	docker tag $(DOCKER_IMAGE_NAME):$(HASH) $(DOCKER_IMAGE_NAME):latest

docker_dev:
	docker build -t $(DOCKER_IMAGE_NAME):$(HASH) .
	docker tag $(DOCKER_IMAGE_NAME):$(HASH) $(DOCKER_IMAGE_NAME):dev

docker_publish:
	make docker
	docker push $(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_IMAGE_NAME):$(HASH)

docker_dev_publish:
	make docker
	docker push $(DOCKER_IMAGE_NAME):dev
	docker push $(DOCKER_IMAGE_NAME):$(HASH)