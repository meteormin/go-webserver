PRJ_NAME=ELK Stack
PRJ_DESC=$(PRJ_NAME) Deployment and Development Makefile
PRJ_BASE=$(shell pwd)

email="sm.yoo@precision-bio.com"

.DEFAULT: help
.SILENT:;

##help: helps (default)
.PHONY: help
help: Makefile
	echo ""
	echo " $(PRJ_DESC)"
	echo ""
	echo " Usage:"
	echo ""
	echo "	make {command}"
	echo ""
	echo " Commands:"
	echo ""
	sed -n 's/^##/	/p' $< | column -t -s ':' |  sed -e 's/^/ /'
	echo ""

##run: run web server
.PHONY: run
run:
	go run ./main.go

##build: build web server
.PHONY: build
build:
	go build -o build/go-webserver ./main.go

##run-docker port={outbound port}: run web server docker
.PHONY: run-docker
run-docker:
	docker run -d --name go-webserver \
		$(if $(port),-p $(port):8080) \
		--network shared_network \
		-v ./web:/go-webserver/web go-webserver

##build-docker: build web server docker
.PHONY: build-docker
build-docker:
	docker build -t go-webserver:latest .
