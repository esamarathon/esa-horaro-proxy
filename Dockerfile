FROM golang:1.13-alpine AS build
RUN apk add --no-cache --virtual .build-deps \
  bash \
  gcc \
  ca-certificates \
  git \
  musl-dev
WORKDIR /go/src/app
COPY ./src .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o webserver .
RUN adduser -S -D -H -h /go/src/app webserver
USER webserver

FROM scratch
LABEL author="Christian Ruigrok <christian@ruigrok.info>"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/app/webserver /app/
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["./webserver"]


