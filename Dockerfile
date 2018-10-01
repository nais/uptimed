FROM golang:1.11-alpine as builder
RUN apk add --no-cache git
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
COPY . /src
WORKDIR /src
RUN rm -f go.sum
RUN go get
RUN go test ./...
RUN go build -a -installsuffix cgo -o uptimed

FROM alpine:3.8
MAINTAINER Sten Røkke <sten.ivar.rokke@nav.no>
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/uptimed /app/uptimed
CMD ["/app/uptimed"]
