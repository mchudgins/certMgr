#
# sample micro-service written in golang using gRPC
#

NAME	:= golang-service-starter
DESC	:= template for golang services
PREFIX	?= usr/local
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER	:= $(shell echo "`git config user.name` <`git config user.email`>")
BUILD_NUMBER_FILE=.buildnum
BUILD_NUM := $(shell if [ -f ${BUILD_NUMBER_FILE} ]; then cat ${BUILD_NUMBER_FILE}; else echo 0; fi)
PKG_RELEASE ?= 1
PROJECT_URL := "git@github.com:mchudgins/golang-service-starter.git"
HYGIENEPKG := "github.com/mchudgins/golang-service-starter/pkg/utils"
LDFLAGS	:= -X '$(HYGIENEPKG).version=$(VERSION)' \
	-X '$(HYGIENEPKG).buildTime=$(BUILDTIME)' \
	-X '$(HYGIENEPKG).builder=$(BUILDER)' \
	-X '$(HYGIENEPKG).goversion=$(GOVERSION)' \
	-X '$(HYGIENEPKG).buildNum=$(BUILD_NUM)'

DEPS := $(shell ls *.go | sed 's/.*_test.go//g')
PROTO_GEN_FILES := pkg/service/service.pb.go \
	pkg/service/common.pb.go \
	pkg/service/service.pb.gw.go

GENERATED_FILES := $(PROTO_GEN_FILES) cmd/bindata.go

.PHONY: fmt test fulltest run container clean site $(BUILD_NUMBER_FILE)

# rule for .pb.gw.go files
%.pb.gw.go: %.proto
	cd pkg/service \
		&& protoc -I/usr/local/include \
			-I. \
 			-I$(GOPATH)/src \
 			-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
 			--grpc-gateway_out=logtostderr=true:. \
 			$(shell basename $<) \
		&& protoc -I/usr/local/include \
			 -I. \
			 -I$(GOPATH)/src \
			 -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
			 --swagger_out=logtostderr=true:. \
			 $(shell basename $<)

%.pb.go: %.proto
	cd pkg/service && protoc -I/usr/local/include -I. \
			 	-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
			 	--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. \
			 	*.proto

all: fmt container

fmt:
	go fmt
#	godep go fix .

build: $(NAME)


cmd/bindata.go: pkg/service/service.pb.gw.go
	@echo The next step will generate a message "(\"no buildable Go source files\")" which may be safely ignored.
	@-go get github.com/swagger-api/swagger-ui
	go-bindata -pkg frontend pkg/service/service.swagger.json $(GOPATH)/src/github.com/swagger-api/swagger-ui/dist
	mv bindata.go pkg/frontend

$(NAME): fmt $(DEPS) $(BUILD_NUMBER_FILE) $(GENERATED_FILES)
	go build -ldflags "$(LDFLAGS)" -o $(NAME)

test: $(DEPS) $(GENERATED_FILES)
	godep go test -v $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

coverage: $(DEPS) $(GENERATED_FILES)
	godep go test -v -coverprofile=cover.out $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)
	godep go tool cover -html=cover.out -o cover.html

fulltest: $(DEPS) $(GENERATED_FILES)
	godep go test -v -cpuprofile=cpu.out
	godep go test -v -blockprofile=block.out
	godep go test -v -memprofile=mem.out

run: $(DEPS) $(BUILD_NUMBER_FILE) $(GENERATED_FILES)
	godep go run -ldflags "$(LDFLAGS)" $(DEPS) -http :9090

container: $(DEPS) docker/Dockerfile $(GENERATED_FILES)
	go get ./...
	godep restore
	CGO_ENABLED=0 godep go build -a -ldflags "$(LDFLAGS) '-s'" -o $(NAME)
	upx -9 -q $(NAME) -o docker/app
	cp $(NAME) docker/app
	docker build -t $(NAME):$(BUILD_NUM) docker

deploy:
	oc new-app --file openshift-deployer-template.json -p APPLICATION=backend,BASE_IMAGESTREAM=scratch,GIT_URI=https://github.com/mchudgins/golang-backend-starter.git

$(BUILD_NUMBER_FILE):
	@if ! test -f $(BUILD_NUMBER_FILE); then echo 0 > $(BUILD_NUMBER_FILE); echo setting file to zero; fi
	@echo $$(($$(cat $(BUILD_NUMBER_FILE)) + 1)) > $(BUILD_NUMBER_FILE)

clean:
	- rm -f certs $(NAME) *.zip *.js *.out docker/app pkg/service/*.go \
			pkg/service/*.json pkg/frontend/bindata.go
