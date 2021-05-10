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

statik-build:
	go get -v github.com/rakyll/statik
	statik -src template -f

go-cli-install: $(GO_SRCS)
	go install ./cmd/aionctl

go-build: $(GO_SRCS)
	go build -o ./dst/service-broker ./cmd/service-broker
	go build -o ./dst/kanban-server ./cmd/kanban-server
	go build -o ./dst/send-anything ./cmd/send-anything
	go build -o ./dst/kanban-replicator ./cmd/kanban-replicator

go-install: $(GO_SRCS)
	statik -src template -f
	go install ./cmd/service-broker
	go install ./cmd/kanban-server
	go install ./cmd/send-anything
	go install ./cmd/kanban-replicator

################## testing
test: go-test

go-mock-build:
	mockgen -source ./config/project.go -destination ./test/mock_config/mock_config.go

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
	protoc --go_out=plugins=grpc,paths=source_relative:./ ./proto/kanbanpb/status.proto
	protoc --go_out=plugins=grpc,paths=source_relative:./ ./proto/devicepb/device.proto
	protoc --go_out=plugins=grpc,paths=source_relative:./ ./proto/servicepb/service.proto
	protoc -I./proto --go_out=plugins=grpc,paths=source_relative:./proto ./proto/clusterpb/cluster.proto
	protoc -I./proto --go_out=plugins=grpc,paths=source_relative:./proto ./proto/projectpb/project.proto




################## initialized grafana (required root permission)
# init-grafana:
# 	mkdir -p /var/lib/aion/grafana
# 	cp ./builders/k8s/prometheus/grafana.db /var/lib/aion/grafana

.PHONY: all, install, build_proto, test
