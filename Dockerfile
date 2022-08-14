FROM golang:1.17-alpine AS build-env

RUN apk add build-base

RUN mkdir /app/
COPY . /app/
WORKDIR /app/

ENV CGO_ENABLED=1
RUN go build -ldflags "-X main.VERSION=$VERSION" github.com/ZamarianPatrick/lazypig-backend

FROM alpine:3.15
COPY --from=build-env /app/lazypig-backend /bin/lazypig

RUN mkdir /app/
WORKDIR /app/

ENTRYPOINT /bin/lazypig