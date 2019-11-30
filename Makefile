# Docker parameters
DOCKER=docker
DOCKERTAG=latest
DOCKERPARAMS=

.PHONY: docker-consumer docker-producer

all: docker-consumer docker-producer

docker-consumer:
	$(DOCKER) build -f worker/consumer/Dockerfile -t $(DOCKERPREFIX)consumer .

docker-producer:
	$(DOCKER) build -f worker/producer/Dockerfile -t $(DOCKERPREFIX)producer .
