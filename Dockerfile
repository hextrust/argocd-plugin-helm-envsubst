FROM golang:1.18-alpine3.16 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG GOOS GOARCH
# CGO_ENABLED=0 for cross platform build
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o argocd-helm-envsubst-plugin

FROM alpine:3.16

WORKDIR /app
COPY --from=builder /app/argocd-helm-envsubst-plugin .
