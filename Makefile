GIT_HASH := $(shell git rev-parse --short HEAD)

build-run-sh:
	docker build -t argocd-helm-envsubst-plugin .
	docker rm -f argocd-helm-envsubst-plugin
	docker run --name argocd-helm-envsubst-plugin -d argocd-helm-envsubst-plugin tail -f /dev/null
	docker exec -it argocd-helm-envsubst-plugin sh

build-push:
	docker buildx build --platform linux/amd64 -t registry.gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin:$(GIT_HASH) .
	docker push registry.gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin:$(GIT_HASH)
	docker buildx build --platform linux/arm64 -t registry.gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin:$(GIT_HASH)-arm64 .
	docker push registry.gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin:$(GIT_HASH)-arm64
