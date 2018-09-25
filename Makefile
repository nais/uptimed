NAME       := uptimed
TAG        := navikt/${NAME}
LATEST     := ${TAG}:latest
GO_IMG     := golang:1.11
GO         := docker run --rm -v ${PWD}:/go/src/github.com/nais/uptimed -w /go/src/github.com/nais/uptimed ${GO_IMG} go


.PHONY: build docker local install docker docker-push linux test

build:
	${GO} build -o uptimed

docker:
	docker image build -t ${TAG}:$(shell /bin/cat ./version) -t ${TAG} -t ${NAME} -t ${LATEST} -f Dockerfile .

docker-push:
	docker image push ${TAG}:$(shell /bin/cat ./version)
	docker image push ${LATEST}

local:
	${GO} run uptimed.go --logtostderr --bind-address=127.0.0.1:8080

install:
	export GO111MODULE=on && ${GO} mod vendor	

test:
	${GO} test ./ ./monitor/
