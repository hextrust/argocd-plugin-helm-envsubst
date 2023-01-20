FROM --platform=$BUILDPLATFORM golang:1.18-alpine3.16 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
# CGO_ENABLED=0 for cross platform build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o argocd-helm-envsubst-plugin

FROM alpine:3.16 as helm-builder

ARG VERSION=3.10.3
ENV BASE_URL="https://get.helm.sh"

WORKDIR /app
RUN case `uname -m` in \
        x86_64) ARCH=amd64; ;; \
        armv7l) ARCH=arm; ;; \
        aarch64) ARCH=arm64; ;; \
        ppc64le) ARCH=ppc64le; ;; \
        s390x) ARCH=s390x; ;; \
        *) echo "un-supported arch, exit ..."; exit 1; ;; \
    esac && \
    apk add --update --no-cache wget git curl && \
    wget ${BASE_URL}/helm-v${VERSION}-linux-${ARCH}.tar.gz -O - | tar -xz && \
    chmod +x linux-${ARCH}/helm && \
    mv linux-${ARCH}/helm /app/helm

FROM alpine:3.16

# Used by plugin to create temporary helm repositories.yaml
RUN mkdir /helm-working-dir 
RUN chmod 777 /helm-working-dir

# Set default helm cache dir to somewhere we can read write
ENV HELM_CACHE_HOME /helm-working-dir

# This is the required location for argocd to recognize the plugin
# https://argo-cd.readthedocs.io/en/stable/user-guide/config-management-plugins/
COPY ConfigManagementPlugin.yaml /home/argocd/cmp-server/config/plugin.yaml

COPY --from=helm-builder /app/helm /usr/bin/

COPY --from=builder /app/argocd-helm-envsubst-plugin /usr/bin/

# Backward compatibility - to be removed
COPY --from=builder /app/argocd-helm-envsubst-plugin /app/argocd-helm-envsubst-plugin
