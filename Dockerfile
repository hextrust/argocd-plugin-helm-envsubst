# Use same image as argocd
FROM ubuntu:21.10

ARG HELM_VERSION="v3.8.1"
ARG HELM_ARH="amd64"
# ARG USER="argocd"
# ARG USER_ID="1001" 

# Base tool
RUN apt update && apt install -y gettext wget

# Install helm
RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-${HELM_ARH}.tar.gz -O - | tar -xz && \
    mv linux-${HELM_ARH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    rm -rf linux-${HELM_ARH}

# Install yq
RUN wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    chmod a+x /usr/local/bin/yq

COPY argocd-helm-envsubst-plugin.sh /usr/bin/argocd-helm-envsubst-plugin
RUN chmod +rx /usr/bin/argocd-helm-envsubst-plugin

# Setup user
# RUN adduser --uid ${USER_ID} ${USER} --disabled-password
# USER ${USER}