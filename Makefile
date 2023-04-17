GIT_HASH := $(shell git rev-parse --short HEAD)

build-run-sh:
	docker build -t argocd-helm-envsubst-plugin .
	docker rm -f argocd-helm-envsubst-plugin
	docker run --name argocd-helm-envsubst-plugin -d argocd-helm-envsubst-plugin tail -f /dev/null
	docker exec -it argocd-helm-envsubst-plugin sh

# GOOS=linux, GOARCH=amd64 works for both minikube and eks
build-push:
	docker buildx create --use
	docker buildx build --platform linux/arm64,linux/amd64 --push --build-arg GOOS=linux --build-arg GOARCH=amd64 -t registry.gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin:$(GIT_HASH) .
	docker buildx rm

build:
	GOOS=darwin GOARCH=amd64 go build -o plugin-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o plugin-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -o plugin-linux-amd64
	GOOS=linux GOARCH=arm64 go build -o plugin-linux-arm64

helm-build:
	go run main.go build --path asset-master-service --repository-path repositories.yaml

helm-render:
	go run main.go render --path asset-master-service

test:
	go test ./... -v