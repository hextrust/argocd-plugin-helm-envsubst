#!/bin/bash

SUPPORTED_ENV='${ARGOCD_ENV_CLUSTER} ${ARGOCD_ENV_ENVIRONMENT} ${ARGOCD_ENV_DOMAIN} ${ARGOCD_ENV_OPERATOR_DOMAIN} ${ARGOCD_ENV_HOSTED_ZONE_ID} ${ARGOCD_ENV_APP_NAME} ${ARGOCD_ENV_ES_HOST} ${ARGOCD_ENV_ES_PORT} ${ARGOCD_ENV_DB_HOST} ${ARGOCD_ENV_DB_PORT} ${ARGOCD_ENV_AWS_ACCOUNT}'

find_helm_config() {
    CONFIG_FILE=""
    if [ -f "config/${ARGOCD_ENV_ENVIRONMENT}.yaml" ]; then
        CONFIG_FILE="-f config/${ARGOCD_ENV_ENVIRONMENT}.yaml"
    fi
}

helm_deploy_crd() {
    SKIP_CRD=$(yq '.argocd.skipCRD' values.yaml)
    HELM_CRD_FLAG="--include-crds"
    if ${SKIP_CRD}; then
        HELM_CRD_FLAG="--skip-crds"
    fi
    echo "SKIP_CRD: ${SKIP_CRD}, HELM_CRD_FLAG: ${HELM_CRD_FLAG}" >> /tmp/argocd-helm-envsubst-plugin.log
}

helm_namespace() {
    NAMESPACE=$(yq '.argocd.namespace' values.yaml | envsubst "$SUPPORTED_ENV")
    HELM_NAMESPACE_FLAG=""
    if [ "${NAMESPACE}" != "null" ]; then
        HELM_NAMESPACE_FLAG="--namespace ${NAMESPACE}"
    fi
}

helm_release_name() {
    RELEASE_NAME=$(yq '.argocd.releaseName' values.yaml | envsubst "$SUPPORTED_ENV")
    HELM_NAME_FLAG=""
    if [ "${RELEASE_NAME}" != "null" ]; then
        HELM_NAME_FLAG="--release-name ${RELEASE_NAME}"
    fi
    echo "RELEASE_NAME: ${RELEASE_NAME}, HELM_NAME_FLAG: ${HELM_NAME_FLAG}" >> /tmp/argocd-helm-envsubst-plugin.log
}

add_sync_option() {
    FILE=$(yq '.argocd.syncOptionReplace.name' values.yaml)
    HELM_POST_RENDER_FLAG=""
    if [ "${FILE}" != "null" ]; then
        # A wrapper script to run kustomize in helm post-renderer
        cat << EOF > kustomize-renderer
#!/bin/bash
cat <&0 > all.yaml
kustomize build . && rm all.yaml && rm kustomization.yaml && rm kustomize-renderer
EOF
        chmod 700 kustomize-renderer

        # Add annotation
        cat << EOF > kustomization.yaml
resources:
- all.yaml
patches:
- patch: |-
    - op: add
      path: /metadata/annotations/argocd.argoproj.io~1sync-options
      value: Replace=true
  target:
    name: $FILE
EOF
        HELM_POST_RENDER_FLAG="--post-renderer ./kustomize-renderer"
    fi
}

logging() {
    LOG_DIR="/tmp/argocd-helm-envsubst-plugin"
    mkdir -p "${LOG_DIR}"
    find "${LOG_DIR}" -mtime +2 -type f -delete
    LOG_FILE_NAME=$(date +%m-%d-%Y).log
    CHART_NAME=$(yq '.name' Chart.yaml)
    echo "[$(date)] [${CHART_NAME}] helm template ${HELM_NAMESPACE_FLAG} ${HELM_CRD_FLAG} ${CONFIG_FILE} ${HELM_NAME_FLAG} ${HELM_POST_RENDER_FLAG} ." >> "${LOG_DIR}/${LOG_FILE_NAME}"
}

find_helm_config
helm_deploy_crd
helm_namespace
helm_release_name
add_sync_option

## Logging
logging

## Run template and envsubst
helm template ${HELM_NAMESPACE_FLAG} ${HELM_CRD_FLAG} ${CONFIG_FILE} ${HELM_NAME_FLAG} ${HELM_POST_RENDER_FLAG} . | envsubst "$SUPPORTED_ENV"
