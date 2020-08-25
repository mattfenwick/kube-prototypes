IMAGE=docker.io/mfenwick100/kube-prototypes-client-server:latest
IMAGE_IP=docker.io/mfenwick100/kube-prototypes-ip-tester:latest

build:
	http-tester
	ip-tester

http-tester:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/http-tester/http-tester ./cmd/http-tester
	docker build -t $(IMAGE) ./cmd/http-tester
	# docker push $(IMAGE)

ip-tester:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/ip-tester/ip-tester ./cmd/ip-tester
	docker build -t $(IMAGE_IP) ./cmd/ip-tester
	# docker push $(IMAGE_IP)

test:
	go test ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...
