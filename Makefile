#
# certMgr, a certificate manager, written in golang using gRPC
#

NAME	:= certMgr
DESC	:= a simple service for generating self-signed certificates
PREFIX	?= usr/local
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER	:= $(shell echo "`git config user.name` <`git config user.email`>")
BUILD_NUMBER_FILE=.buildnum
BUILD_NUM := $(shell if [ -f ${BUILD_NUMBER_FILE} ]; then cat ${BUILD_NUMBER_FILE}; else echo 0; fi)
PKG_RELEASE ?= 1
PROJECT_URL := "git@github.com:mchudgins/certMgr.git"
HYGIENEPKG := "github.com/mchudgins/certMgr/pkg/utils"
LDFLAGS	:= -X '$(HYGIENEPKG).version=$(VERSION)' \
	-X '$(HYGIENEPKG).buildTime=$(BUILDTIME)' \
	-X '$(HYGIENEPKG).builder=$(BUILDER)' \
	-X '$(HYGIENEPKG).goversion=$(GOVERSION)' \
	-X '$(HYGIENEPKG).buildNum=$(BUILD_NUM)'

DEPS := $(shell ls *.go | sed 's/.*_test.go//g')
PROTO_GEN_FILES := pkg/service/service.pb.go \
	pkg/service/common.pb.go \
	pkg/service/service.pb.gw.go \
	pkg/service/certMgrService.pb.go \
	pkg/service/certMgrService.pb.gw.go

GENERATED_FILES := $(PROTO_GEN_FILES) pkg/assets/bindata_assetfs.go pkg/frontend/bindata.go ui/site/index.html

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

build: $(NAME)

pkg/frontend/bindata.go: pkg/service/service.pb.gw.go
	@echo The next step will generate a message "(\"no buildable Go source files\")" which may be safely ignored.
	@-go get github.com/swagger-api/swagger-ui
	go-bindata -pkg frontend pkg/service/service.swagger.json $(GOPATH)/src/github.com/swagger-api/swagger-ui/dist
	mv bindata.go pkg/frontend

pkg/assets/bindata_assetfs.go: ui/site/index.html
	mkdir -p pkg/assets
	go-bindata-assetfs -pkg assets -prefix ui/site ui/site/...
	mv bindata_assetfs.go pkg/assets

ui/site/index.html: ui/src/homepage.html
	cd ui && make

$(NAME): fmt $(DEPS) $(BUILD_NUMBER_FILE) $(GENERATED_FILES)
	go install -ldflags "$(LDFLAGS)"

test: $(DEPS) $(GENERATED_FILES)
	go test -v $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

coverage: $(DEPS) $(GENERATED_FILES)
	go test -v -coverprofile=cover.out $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)
	go tool cover -html=cover.out -o cover.html

fulltest: $(DEPS) $(GENERATED_FILES)
	go test -v -cpuprofile=cpu.out
	go test -v -blockprofile=block.out
	go test -v -memprofile=mem.out

run: $(DEPS) $(BUILD_NUMBER_FILE) $(GENERATED_FILES)
	go run -ldflags "$(LDFLAGS)" $(DEPS) backend --http :9090

container: $(DEPS) docker/Dockerfile $(GENERATED_FILES)
	CGO_ENABLED=0 go build -a -ldflags "$(LDFLAGS) '-s'" -o bin/$(NAME)
	@-rm docker/app
	upx -9 -q bin/$(NAME) -o docker/app
	cp bin/$(NAME) docker/app
	docker build -t cert-mgr:$(BUILD_NUM) docker

deploy:
	-oc secrets new certmgrkeys ca-key.pem=ca/cap/private/cap-ca.key key.pem=key.pem
	oc new-app --file openshift-deployer-template.json \
	    -p APPLICATION=certmgr \
	    -p GIT_URI=https://github.com/mchudgins/certMgr.git
	oc start-build certmgr

undeploy:
	oc process -f openshift-deployer-template.json \
	    -v APPLICATION=certmgr \
	    -v GIT_URI=https://github.com/mchudgins/certMgr.git | oc delete -f -

$(BUILD_NUMBER_FILE):
	@if ! test -f $(BUILD_NUMBER_FILE); then echo 0 > $(BUILD_NUMBER_FILE); echo setting file to zero; fi
	@echo $$(($$(cat $(BUILD_NUMBER_FILE)) + 1)) > $(BUILD_NUMBER_FILE)

clean:
	- rm -f certs $(NAME) *.zip *.js *.out docker/app pkg/service/*.go \
			pkg/service/*.json pkg/frontend/bindata.go
