package internal

import (
	"testing"
	// "gopkg.in/yaml.v2"
	// "io/ioutil"
	"log"
	"path"
    "path/filepath"
    "runtime"
	"os"
	// "os/exec"
	// "strings"
)

type ChartYaml struct {
	
}

/*
asset-master-service/Chart.yaml
apiVersion: v2
name: asset-master-service
description: Asset master
dependencies:
- name: generic-helm-chart
  version: 0.2.15
  repository: https://gitlab.int.hextech.io/api/v4/projects/645/packages/helm/stable
type: application
version: 1.0.0
appVersion: "1.0.0"
*/

func TestBuildPrivateRepo(t *testing.T) {
	basePath := getProjectRoot()
	log.Printf("basePath: %s", basePath)
	createTempFolder(basePath)
	createChartYaml(basePath)
	

	// const helmChartPath = "_tmp"
	// const repoConfigPath = ""
	// NewBuilder().Build(helmChartPath, repoConfigPath)
}

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
    d := path.Join(path.Dir(b))
    return filepath.Dir(d)
}

func createTempFolder(basePath string) {
	os.Chdir(basePath)
	if _, err := os.Stat(basePath + "/" + "_tmp"); os.IsNotExist(err) {
		log.Println("_tmp folder does not exist.")
	} else {
		log.Println("_tmp folder exists.")
	}	
}

func createChartYaml() {

}
