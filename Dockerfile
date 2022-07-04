FROM alpine:3.16.0

ARG HELM_VERSION="v3.8.1"
ARG HELM_ARH="amd64"
# Name "argocd" is required by argocd
ARG USER="argocd"
# ID "999" is required by argocd
ARG USER_ID="999" 

# Base tool
RUN apk update && apk add curl wget gettext yq --no-cache --update

# Install helm
RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-${HELM_ARH}.tar.gz -O - | tar -xz && \
    mv linux-${HELM_ARH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    rm -rf linux-${HELM_ARH}

COPY argocd-helm-envsubst-plugin.sh /usr/bin/argocd-helm-envsubst-plugin

# Move existing group to give room for id 999
RUN sed -i "s/999/99/" /etc/group
RUN adduser -h /home/${USER} -s /usr/bin/bash -u ${USER_ID} ${USER} -D
USER ${USER}

