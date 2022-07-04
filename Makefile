build-run-sh:
	docker build -t argocd-helm-envsubst-plugin .
	docker rm -f argocd-helm-envsubst-plugin
	docker run --name argocd-helm-envsubst-plugin -d argocd-helm-envsubst-plugin tail -f /dev/null
	docker exec -it argocd-helm-envsubst-plugin sh