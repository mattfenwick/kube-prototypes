IMAGE=docker.io/mfenwick100/kube-prototypes-client-server:latest

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/http-tester/http-tester ./cmd/http-tester
	docker build -t $(IMAGE) ./cmd/http-tester
	# docker push $(IMAGE)

test:
	go test ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...
