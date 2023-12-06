#------ Build golang app ------#
FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.18 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
# CGO_ENABLED=0 for cross platform build
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o argocd-helm-envsubst-plugin

#------ Install dependening software ------#
FROM alpine:3.18 as helm-builder

# amd64/arm64
ARG TARGETARCH
WORKDIR /app
RUN apk add --update --no-cache wget git curl

# Install helm
ARG HELM_VERSION=3.10.3
ENV HELM_BASE_URL="https://get.helm.sh"
RUN wget ${HELM_BASE_URL}/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz -O - | tar -xz && \
    chmod +x linux-${TARGETARCH}/helm && \
    mv linux-${TARGETARCH}/helm /app/helm

# Install kustomize
ARG KUSTOMIZE_VERSION=4.5.7
ENV KUSTOMIZE_BASE_URL="https://github.com/kubernetes-sigs/kustomize/releases/download"
RUN wget ${KUSTOMIZE_BASE_URL}/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${TARGETARCH}.tar.gz -O - | tar -xz && \
    chmod +x kustomize

#------ Final image ------# 
FROM alpine:3.18

# Used by plugin to create temporary helm repositories.yaml
RUN mkdir /helm-working-dir 
RUN chmod 777 /helm-working-dir

# Set default helm cache dir to somewhere we can read write
ENV HELM_CACHE_HOME /helm-working-dir

# This is the required location for argocd to recognize the plugin
# https://argo-cd.readthedocs.io/en/stable/user-guide/config-management-plugins/
COPY ConfigManagementPlugin.yaml /home/argocd/cmp-server/config/plugin.yaml

COPY --from=helm-builder /app/helm /usr/bin/
COPY --from=helm-builder /app/kustomize /usr/bin/
COPY --from=builder /app/argocd-helm-envsubst-plugin /usr/bin/

# Backward compatibility - to be removed
COPY --from=builder /app/argocd-helm-envsubst-plugin /app/argocd-helm-envsubst-plugin
