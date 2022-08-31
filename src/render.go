package src

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	envsubstList = []string{
		"ARGOCD_ENV_CLUSTER",
		"ARGOCD_ENV_ENVIRONMENT",
		"ARGOCD_ENV_DOMAIN",
		"ARGOCD_ENV_OPERATOR_DOMAIN",
		"ARGOCD_ENV_HOSTED_ZONE_ID",
		"ARGOCD_ENV_APP_NAME",
		"ARGOCD_ENV_ES_HOST",
		"ARGOCD_ENV_ES_PORT",
		"ARGOCD_ENV_DB_HOST",
		"ARGOCD_ENV_DB_PORT",
		"ARGOCD_ENV_AWS_ACCOUNT",
	}

	defaultDebugLogFilePath = "/tmp/argocd-helm-envsubst-plugin/"
	defaultHelmChartPath    = "./"
)

type HelmConfig struct {
	ArgocdConfig HexArgocdPluginConfig `yaml:"argocd,omitempty"`
}

type HexArgocdPluginConfig struct {
	ReleaseName       string   `yaml:"releaseName,omitempty"`
	Namespace         string   `yaml:"namespace,omitempty"`
	SkipCRD           bool     `yaml:"skipCRD,omitempty"`
	SyncOptionReplace []string `yaml:"syncOptionReplace,omitempty"`
}

func RenderTemplate(helmChartPath string, debugLogPath string) {
	if len(debugLogPath) <= 0 {
		debugLogPath = defaultDebugLogFilePath
	}

	if len(helmChartPath) <= 0 {
		helmChartPath = defaultHelmChartPath
	}

	os.Chdir(helmChartPath)

	command := "helm"
	args := []string{"template"}

	configFileName := findHelmConfig()
	if len(configFileName) > 0 {
		args = append(args, "-f")
		args = append(args, configFileName)
	}

	configFile := mergeYaml("values.yaml", configFileName)
	argocdConfig := readArgocdConfig(configFile)

	if len(argocdConfig.Namespace) > 0 {
		args = append(args, "--namespace")
		args = append(args, argocdConfig.Namespace)
	}

	if len(argocdConfig.ReleaseName) > 0 {
		args = append(args, "--release-name")
		args = append(args, argocdConfig.ReleaseName)
	}

	if argocdConfig.SkipCRD {
		args = append(args, "--skip-crds")
	} else {
		args = append(args, "--include-crds")
	}

	if len(argocdConfig.SyncOptionReplace) > 0 {
		postRendererScript := preparePostRenderer(argocdConfig.SyncOptionReplace)
		args = append(args, "--post-renderer")
		args = append(args, postRendererScript)
	}

	args = append(args, ".")
	cmd := envsubst(strings.Join(args, " "))
	debugLog(cmd+"\n", debugLogPath)

	out, err := exec.Command(command, strings.Split(cmd, " ")...).Output()
	if err != nil {
		log.Fatalf("Exec helm template error: %v", err)
	}

	manifest := envsubst(string(out))
	fmt.Println(manifest)
}

func dependencyBuild() {
	out, err := exec.Command("helm", "dependency", "build").Output()
	if err != nil {
		log.Fatalf("Exec helm dependency build error: %v", err)
	}
	log.Printf("%s\n", out)
}

func findHelmConfig() string {
	var files []string
	root := "config/"
	environment := os.Getenv("ARGOCD_ENV_ENVIRONMENT")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// log.Fatalf("Config folder not found: %v", err)
			return nil
		}
		if !info.IsDir() && path == root+environment+".yaml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Find config file in dir error: %v", err)
	}
	if len(files) > 0 {
		return files[0]
	}
	return ""
}

func readArgocdConfig(configFile string) *HexArgocdPluginConfig {
	c := HelmConfig{}
	err := yaml.Unmarshal([]byte(configFile), &c)
	if err != nil {
		log.Fatalf("Unmarshal config file error: %v", err)
	}
	return &c.ArgocdConfig
}

func envsubst(str string) string {
	for _, env := range envsubstList {
		envVar := os.Getenv(env)
		if len(envVar) > 0 {
			str = strings.Replace(str, "${"+env+"}", envVar, -1)
		}
	}
	return str
}

func preparePostRenderer(files []string) string {
	scriptFileName := "./kustomize-renderer"

	// Create shell script
	script := `#!/bin/bash
cat <&0 > all.yaml
kustomize build . && rm all.yaml && rm kustomization.yaml && rm kustomize-renderer`

	err := os.WriteFile(scriptFileName, []byte(script), 0777)
	if err != nil {
		log.Fatalf("Create kustomize-renderer error: %s", err)
	}

	// Create kustomize file
	kustomizations := []string{fmt.Sprintf(
		"resources:\n" +
			"- all.yaml\n" +
			"patches:")}

	for _, file := range files {
		kustomizations = append(kustomizations, fmt.Sprintf(
			"- patch: |-\n"+
				"    - op: add\n"+
				"      path: /metadata/annotations/argocd.argoproj.io~1sync-options\n"+
				"      value: Replace=true\n"+
				"  target:\n"+
				"    name: %v", file))
	}

	err = os.WriteFile("./kustomization.yaml", []byte(strings.Join(kustomizations, "\n")), 0777)
	if err != nil {
		log.Fatalf("Create kustomization.yaml error: %s", err)
	}

	return scriptFileName
}

func mergeYaml(filenames ...string) string {
	if len(filenames) <= 0 {
		log.Fatalf("You must provide at least one filename for reading Values")
	}
	var resultValues map[string]interface{}
	for _, filename := range filenames {

		var override map[string]interface{}
		bs, err := ioutil.ReadFile(filename)
		if err != nil {
			// log.Println(err)
			continue
		}
		if err := yaml.Unmarshal(bs, &override); err != nil {
			// log.Println(err)
			continue
		}

		//check if is nil. This will only happen for the first filename
		if resultValues == nil {
			resultValues = override
		} else {
			for k, v := range override {
				resultValues[k] = v
			}
		}

	}
	bs, err := yaml.Marshal(resultValues)
	if err != nil {
		log.Fatalf("Marshal file error:", err)
	}

	return string(bs)
}

func debugLog(cmd string, debugLogFilePath string) {
	date := time.Now()
	formattedDate := date.Format("01-02-2006")
	logFilePath := debugLogFilePath + formattedDate + ".log"

	// Log line name - Get from Chart.yaml name field
	chartYaml := readChartYaml()

	// Create log line
	formattedDateTime := date.Format("2006-01-02 15:04:05.000000")
	logLine := fmt.Sprintf("[%s][%s] %s", formattedDateTime, chartYaml["name"], cmd)

	// Create log file if not exist
	if _, err := ioutil.ReadFile(logFilePath); err != nil {
		err = os.Mkdir(debugLogFilePath, 0755)
		if err != nil {
			log.Fatalf("Failed to create log folder: %v", err)
		}

		f, err := os.Create(logFilePath)
		if err != nil {
			log.Fatalf("Fail to create log file: %v", err)
		}
		f.Close()
	}

	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("Fail to opn debug log file:", err)
	}

	// Write log line
	if _, err = f.WriteString(logLine); err != nil {
		log.Fatalf("Fail to write debug log file:", err)
	}
}

func readChartYaml() map[string]interface{} {
	var chartYaml map[string]interface{}
	bs, err := ioutil.ReadFile("Chart.yaml")
	if err != nil {
		log.Fatalf("Read Chart.yaml error: %v", err)
	}
	if err := yaml.Unmarshal(bs, &chartYaml); err != nil {
		log.Fatalf("Unmarshal Chart.yaml error: %v", err)
	}
	return chartYaml
}
