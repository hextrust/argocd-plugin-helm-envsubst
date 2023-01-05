#!/bin/bash

# echo "run..."

export ARGOCD_ENV_ENVIRONMENT=alpha
export ARGOCD_ENV_ES_HOST=clustercfg.elasticache-metazen-beta.fckkzb.apse1.cache.amazonaws.com 
export ARGOCD_ENV_ES_PORT=6379
export ARGOCD_ENV_CHAIN_ID=4
export ARGOCD_ENV_CLUSTER=metazen
export ARGOCD_ENV_DOMAIN=test.alpha.metazens.xyz
export ARGOCD_ENV_OPERATOR_DOMAIN=develop.operator.hextech.io
export ARGOCD_ENV_HOSTED_ZONE_ID=Z08321952NOP80HU1BX1D
export ARGOCD_ENV_DB_HOST=postgres-metazen-beta.cag9qu9vg30h.ap-southeast-1.rds.amazonaws.com
export ARGOCD_ENV_DB_PORT=5432
export ARGOCD_ENV_CONFIRMATION_COUNT=10


# Rinkeby
export ARGOCD_ENV_APP_NAME=blockchain-listener-ethereum-rinkeby
export ARGOCD_ENV_TOKEN=ethereum
export ARGOCD_ENV_TICKER=ETH
export ARGOCD_ENV_NETWORK=rinkeby
export ARGOCD_ENV_GETH_ENDPOINT=http://54.255.28.145:8545

# mumbai
# export ARGOCD_ENV_APP_NAME=blockchain-listener-polygon-mumbai
# export ARGOCD_ENV_TOKEN=polygon
# export ARGOCD_ENV_TICKER=MATIC
# export ARGOCD_ENV_NETWORK=mumbai
# export ARGOCD_ENV_GETH_ENDPOINT=https://small-green-breeze.matic-testnet.quiknode.pro/fbd755dbab8669355c48e512cee494083ff6673f

# env | grep ARGOCD_ENV 
APP=_tmp/asset-master-service
go run main.go render --path ${APP}
