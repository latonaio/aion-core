################### build
all: go-build ;
install: go-install ;

GO_SRCS := $(shell find . -type f -name '*.go')

docker-build: $(GO_SRCS)
	bash ./builders/docker-build-servicebroker.sh
	bash ./builders/docker-build-statuskanban.sh
	bash ./builders/docker-build-sendanything.sh
	bash ./builders/docker-build-kanban-replicator.sh

docker-push:
	bash ./builders/docker-build-servicebroker.sh push
	bash ./builders/docker-build-statuskanban.sh push
	bash ./builders/docker-build-sendanything.sh push
	bash ./builders/docker-build-kanban-replicator.sh push

required-package:
	go get -v github.com/rakyll/statik
	statik -src template -f

go-build: $(GO_SRCS)
	go build ./cmd/service-broker
	go build ./cmd/kanban-server
	go build ./cmd/send-anything
	go build ./cmd/kanban-replicator

go-install: $(GO_SRCS)
	statik -src template -f
	go install ./cmd/service-broker
	go install ./cmd/kanban-server
	go install ./cmd/send-anything
	go install ./cmd/kanban-replicator

################## testing
test: go-test python-test

go-mock-build:
	mockgen -source ./config/project.go -destination ./test/mock_config/mock_config.go
	mockgen -source ./pkg/wsclient/client.go -destination ./test/mock_wsclient/mock_wsclient.go

GO_PACKAGES=`go list ./... | grep -v -e test -e sftp`
go-test: go-mock-build
	go test $(GO_PACKAGES) -v -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html

################## build proto
build_proto: python-proto go-proto

PY_PROTO_DIR = ./python/aion/proto
python-proto: proto
	cp ./proto/kanbanpb/*.proto $(PY_PROTO_DIR)
	cd ./python && \
		python3 -m grpc_tools.protoc -I./ \
		--python_out=. \
		--grpc_python_out=. \
		./aion/proto/*.proto
	rm $(PY_PROTO_DIR)/*.proto

go-proto: proto
	protoc --go_out=plugins:. ./proto/kanbanpb/status.proto

################## initialized grafana (required root permission)
# init-grafana:
# 	mkdir -p /var/lib/aion/grafana
# 	cp ./builders/k8s/prometheus/grafana.db /var/lib/aion/grafana

.PHONY: all, install, build_proto, test