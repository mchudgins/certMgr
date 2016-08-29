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
BUILD_NUMBER_FILE=.buildnum
BUILD_NUM := $(shell if [ -f ${BUILD_NUMBER_FILE} ]; then cat ${BUILD_NUMBER_FILE}; else echo 0; fi)
PKG_RELEASE ?= 1
PROJECT_URL := "git@svn.dstresearch.com:devOps/certManager"
LDFLAGS	:= -X 'main.version=$(VERSION)' \
	-X 'main.buildTime=$(BUILDTIME)' \
	-X 'main.builder=$(BUILDER)' \
	-X 'main.goversion=$(GOVERSION)' \
	-X 'main.buildNum=$(BUILD_NUM)'

PB_DEPS := ../service/service.pb.go ../service/common.pb.go

DEPS := $(shell ls *.go | sed 's/.*_test.go//g')

.PHONY: fmt test fulltest run container clean site $(BUILD_NUMBER_FILE)

all: fmt container

fmt:
	go fmt
#	godep go fix .

build: $(NAME)

$(PB_DEPS): ../service/service.proto ../service/common.proto
	cd ../service \
		&& protoc -I/usr/local/include -I. \
			-I$(GOPATH)/src \
			-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
			--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. \
			*.proto

$(NAME): fmt $(DEPS) $(BUILD_NUMBER_FILE) $(PB_DEPS)
	godep go build -ldflags "$(LDFLAGS)" -o $(NAME)

test: $(DEPS)
	godep go test -v $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

coverage: $(DEPS)
	godep go test -v -coverprofile=cover.out $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)
	godep go tool cover -html=cover.out -o cover.html

fulltest: $(DEPS)
	godep go test -v -cpuprofile=cpu.out
	godep go test -v -blockprofile=block.out
	godep go test -v -memprofile=mem.out

run: $(DEPS) $(BUILD_NUMBER_FILE)
	godep go run -ldflags "$(LDFLAGS)" $^ -http :9090

container: $(DEPS) docker/Dockerfile $(PB_DEPS)
	go get ./...
	godep restore
	CGO_ENABLED=0 godep go build -a -ldflags "$(LDFLAGS) '-s'" -o $(NAME)
	upx -9 -q $(NAME) -o docker/app
	cp $(NAME) docker/app
	docker build -t $(NAME):$(BUILD_NUM) docker

deploy:
	oc new-app --file openshift-deployer-template.json -p APPLICATION=backend,BASE_IMAGESTREAM=scratch,GIT_URI=https://github.com/mchudgins/golang-backend-starter.git

run_container:
	#sudo docker run --name mysql -e MYSQL_USER=certs -e MYSQL_PASSWORD=certs -e MYSQL_DATABASE=certs -e MYSQL_ROOT_PASSWORD=password -d mysql
	sudo docker run -it --rm -p 9443:8443 -p 9444:8444 -e DB="certs:certs@tcp(mysql:3306)/certs?parseTime=true" --link mysql certs

$(BUILD_NUMBER_FILE):
	@if ! test -f $(BUILD_NUMBER_FILE); then echo 0 > $(BUILD_NUMBER_FILE); echo setting file to zero; fi
	@echo $$(($$(cat $(BUILD_NUMBER_FILE)) + 1)) > $(BUILD_NUMBER_FILE)

clean:
	- rm -f certs *.zip *.js *.out docker/app ../service/*.pb.go
