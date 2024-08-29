FROM alpine AS base

LABEL maintainer="PB-DHC"

ARG TIME_ZONE
ARG TAG

RUN apk --no-cache add tzdata && \
	cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime && \
	echo "${TIME_ZONE}" > /etc/timezone \
	apk del tzdata

FROM golang:1.22-alpine AS build

RUN apk add --no-cache make

WORKDIR /go-webserver

COPY . .

RUN CGO_ENABLED=0 go mod download

RUN make build

FROM base AS  deploy

WORKDIR /go-webserver

COPY --from=build /go-webserver/build/ .

EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./go-webserver"]