package internal_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin/internal"
)

const (
	TEST_ARGOCD_ENV_ENVIRONMENT = "prod"
	TEST_ARGOCD_ENV_CLUSTER     = "blockchain"
)

func setup(t *testing.T) {
	if err := os.Mkdir("config", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Create(fmt.Sprintf("config/%s_%s.yaml", TEST_ARGOCD_ENV_CLUSTER, TEST_ARGOCD_ENV_ENVIRONMENT)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Create(fmt.Sprintf("config/%s.yaml", TEST_ARGOCD_ENV_ENVIRONMENT)); err != nil {
		t.Fatal(err)
	}

	os.Setenv("ARGOCD_ENV_ENVIRONMENT", TEST_ARGOCD_ENV_ENVIRONMENT)
	os.Setenv("ARGOCD_ENV_CLUSTER", TEST_ARGOCD_ENV_CLUSTER)
}

func teardown(t *testing.T) {
	if err := os.RemoveAll("config"); err != nil {
		t.Fatal(err)
	}
}

func TestFindHelmConfigs(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := internal.NewRenderer()

	expect := []string{
		"values.yaml",
		fmt.Sprintf("config/%s.yaml", TEST_ARGOCD_ENV_ENVIRONMENT),
		fmt.Sprintf("config/%s_%s.yaml", TEST_ARGOCD_ENV_CLUSTER, TEST_ARGOCD_ENV_ENVIRONMENT),
	}
	actual := r.FindHelmConfigs()

	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("Expected %s do not match actual %s", expect, actual)
	}
}
