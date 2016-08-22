#
# trivial reverse proxy
#

NAME	:= golang-backend-starter
DESC	:= template for backends running in k8s/openshift
PREFIX	?= usr/local
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER	:= $(shell echo "`git config user.name` <`git config user.email`>")
BUILD_NUMBER_FILE=build.num
BUILD_NUM := $(shell if [ -f build.num ]; then cat build.num; else echo 1 >build.num && echo 1; fi)
PKG_RELEASE ?= 1
PROJECT_URL := "git@svn.dstresearch.com:devOps/certManager"
LDFLAGS	:= -X 'main.version=$(VERSION)' \
	-X 'main.buildTime=$(BUILDTIME)' \
	-X 'main.builder=$(BUILDER)' \
	-X 'main.goversion=$(GOVERSION)' \
	-X 'main.buildNum=$(BUILD_NUM)'


DEPS := $(shell ls *.go)


.PHONY: fmt test fulltest run container clean site $(BUILD_NUMBER_FILE)

all: fmt container

fmt:
	go fmt
#	godep go fix .

build: $(NAME)

$(NAME): fmt $(DEPS) $(BUILD_NUMBER_FILE)
	godep go build -ldflags "$(LDFLAGS)" -o $(NAME)

test: $(DEPS)
	godep go test -v -coverprofile=cover.out $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)
	godep go tool cover -html=cover.out -o cover.html

fulltest: $(DEPS)
	godep go test -v -cpuprofile=cpu.out
	godep go test -v -blockprofile=block.out
	godep go test -v -memprofile=mem.out

run: $(NAME)
	./$(NAME) -listen :80 -target http://localhost:8000

container: $(DEPS) docker/Dockerfile
	CGO_ENABLED=0 godep go build -a -ldflags "$(LDFLAGS) '-s'" -o $(NAME)
	#upx -9 -q $(NAME) -o docker/app
	cp $(NAME) docker/app
	docker build -t $(NAME):$(BUILD_NUM) docker

run_container:
	#sudo docker run --name mysql -e MYSQL_USER=certs -e MYSQL_PASSWORD=certs -e MYSQL_DATABASE=certs -e MYSQL_ROOT_PASSWORD=password -d mysql
	sudo docker run -it --rm -p 9443:8443 -p 9444:8444 -e DB="certs:certs@tcp(mysql:3306)/certs?parseTime=true" --link mysql certs

$(BUILD_NUMBER_FILE):
	@if ! test -f $(BUILD_NUMBER_FILE); then echo 0 > $(BUILD_NUMBER_FILE); echo setting file to zero; fi
	@echo $$(($$(cat $(BUILD_NUMBER_FILE)) + 1)) > $(BUILD_NUMBER_FILE)

clean:
	- rm -f certs *.zip *.js *.out docker/app
