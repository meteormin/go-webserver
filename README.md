# Go simple static web server

- Go Built-in http패키지를 활용한 정적 웹 서버

## How to use

### 1. Start Go Web Server

```bash
$ make start env=default compose=go-webserver
```

### 2. Check Go Web Server

* Go Web Server: http://localhost:8080

## How to build

```bash
$ make build env=default compose=go-webserver
```

## Docker

```bash
$ make build-docker
$ make run-docker
```
