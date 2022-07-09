 #!/bin/sh

SUPPORTED_ENV='${ARGOCD_ENV_CLUSTER} ${ARGOCD_ENV_ENVIRONMENT} ${ARGOCD_ENV_DOMAIN} ${ARGOCD_ENV_OPERATOR_DOMAIN} ${ARGOCD_ENV_HOSTED_ZONE_ID} ${ARGOCD_ENV_APP_NAME} ${ARGOCD_ENV_ES_HOST} ${ARGOCD_ENV_ES_PORT} ${ARGOCD_ENV_DB_HOST} ${ARGOCD_ENV_DB_PORT}'

## DEBUG mode, default false
if [ ${DEBUG:=0} == 1 ]; then
    set -x
    env
fi

## Get namespacec from value.yaml
NAMESPACE=$(cat values.yaml | sed -n 's/.*kubeNamespace: \(.*\)$/\1/p')

## Find environment config yaml
CONFIG_FILE=""
if [ -f "config/${ARGOCD_ENV_ENVIRONMENT}.yaml" ]; then
    CONFIG_FILE="-f config/${ARGOCD_ENV_ENVIRONMENT}.yaml"
fi

## Run template and envsubst
helm template --namespace $NAMESPACE --include-crds $CONFIG_FILE . | envsubst "$SUPPORTED_ENV"
