package internal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"regexp"

	"gopkg.in/yaml.v2"
)

var (
	defaultDebugLogFilePath = "/tmp/argocd-helm-envsubst-plugin/"
	defaultHelmChartPath    = "./"
	argocdEnvVarPrefix      = "ARGOCD_ENV"
)

type ConfigFileSeq struct {
	Seq  int
	Name string
}

type Renderer struct {
	debugLogFilePath string
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (renderer *Renderer) RenderTemplate(helmChartPath string, debugLogFilePath string) {
	if len(debugLogFilePath) <= 0 {
		renderer.debugLogFilePath = defaultDebugLogFilePath
	} else {
		renderer.debugLogFilePath = debugLogFilePath
	}

	if len(helmChartPath) <= 0 {
		helmChartPath = defaultHelmChartPath
	}

	os.Chdir(helmChartPath)
	envs := renderer.getArgocdEnvList()

	command := "helm"
	args := []string{"template"}

	useExternalHelmChartPathIfSet()

	configFileNames := renderer.FindHelmConfigs()
	if len(configFileNames) > 0 {
		for _, name := range configFileNames {
			args = append(args, "-f")
			args = append(args, name)
			renderer.inlineEnvsubst(name, envs)
		}
	}

	helmConfig := renderer.mergeYaml(configFileNames)
	argocdConfig := ReadArgocdConfig(helmConfig)

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
		postRendererScript := renderer.preparePostRenderer(argocdConfig.SyncOptionReplace)
		args = append(args, "--post-renderer")
		args = append(args, postRendererScript)
	}

	args = append(args, ".")
	strCmd := strings.Join(args, " ")
	renderer.debugLog(strCmd + "\n")

	cmd := exec.Command(command, strings.Split(strCmd, " ")...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Exec helm template error: %s\n%s", err, stderr.String())
	}

	fmt.Println(out.String())
}

func (renderer *Renderer) FindHelmConfigs() []string {
	// Default to values.yaml
	files := []ConfigFileSeq{
		{
			Seq:  0,
			Name: "values.yaml",
		},
	}
	root := "config/"
	environment := os.Getenv("ARGOCD_ENV_ENVIRONMENT")
	cluster := os.Getenv("ARGOCD_ENV_CLUSTER")
	helmConfigFilePatterns := []string{
		"values.yaml",
		root + environment + ".yaml",
		root + cluster + "_" + environment + ".yaml",
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// log.Fatalf("Config folder not found: %v", err)
			return nil
		}
		for seq, pattern := range helmConfigFilePatterns {
			if match, _ := regexp.MatchString(pattern, path); match {
				files = append(files, ConfigFileSeq{Seq: seq, Name: path})
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Find config file in dir error: %v", err)
	}

	// sort
	sort.Slice(files, func(i, j int) bool {
		return files[i].Seq < files[j].Seq
	})

	// convert array of struct to array of string
	sortedFiles := []string{}
	for _, file := range files {
		sortedFiles = append(sortedFiles, file.Name)
	}

	return sortedFiles
}

func (renderer *Renderer) getArgocdEnvList() []string {
	envs := []string{}
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		if strings.HasPrefix(key, argocdEnvVarPrefix) {
			envs = append(envs, key)
		}
	}
	return envs
}

func (renderer *Renderer) inlineEnvsubst(filename string, envs []string) {
	// Read file
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		renderer.debugLog(fmt.Sprintf("inlineEnvsubst - Read file error %v \n", err))
		return
	}
	helmConfig := string(bs)

	// Substitute
	envsubstHelmConfig := renderer.envsubst(helmConfig, envs)

	// Write file
	if err := ioutil.WriteFile(filename, []byte(envsubstHelmConfig), 0644); err != nil {
		renderer.debugLog(fmt.Sprintf("inlineEnvsubst - Write file error %v \n", err))
	}
}

func (renderer *Renderer) envsubst(str string, envs []string) string {
	for _, env := range envs {
		envVar := os.Getenv(env)
		if len(envVar) > 0 {
			str = strings.Replace(str, "${"+env+"}", envVar, -1)
		}
	}
	return str
}

func (renderer *Renderer) preparePostRenderer(files []string) string {
	// Get the current temp path
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("osGetwd error: %s", err)
	}

	scriptPath := pwd + "/kustomize-renderer"
	kustomizeYamlPath := pwd + "/kustomization.yaml"
	allPath := pwd + "/all.yaml"

	// Create shell script
	script := fmt.Sprintf(`#!/bin/sh
	cat <&0 > %s
	kustomize build .`, allPath)

	err = os.WriteFile(scriptPath, []byte(script), 0777)
	if err != nil {
		log.Fatalf("Create kustomize-renderer error: %s", err)
	}

	// Create kustomize file
	kustomizations := []string{fmt.Sprintf(
		"resources:\n"+
			"- %s\n"+
			"patches:", allPath)}

	for _, file := range files {
		kustomizations = append(kustomizations, fmt.Sprintf(
			"- patch: |-\n"+
				"    - op: add\n"+
				"      path: /metadata/annotations/argocd.argoproj.io~1sync-options\n"+
				"      value: Replace=true\n"+
				"  target:\n"+
				"    name: %v", file))
	}

	err = os.WriteFile(kustomizeYamlPath, []byte(strings.Join(kustomizations, "\n")), 0777)
	if err != nil {
		log.Fatalf("Create %s error: %s", kustomizeYamlPath, err)
	}

	return scriptPath
}

func (renderer *Renderer) mergeYaml(configFiles []string) string {
	if len(configFiles) <= 0 {
		log.Fatalf("You must provide at least one config yaml")
	}
	var resultValues map[string]interface{}
	for _, filename := range configFiles {

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
		log.Fatalf("Marshal file error: %v", err)
	}

	return string(bs)
}

func (renderer *Renderer) debugLog(cmd string) {
	date := time.Now()
	formattedDate := date.Format("01-02-2006")
	logFilePath := renderer.debugLogFilePath + formattedDate + ".log"

	// Log line name - Get from Chart.yaml name field
	chartYaml := ReadChartYaml()

	// Create log line
	formattedDateTime := date.Format("2006-01-02 15:04:05.000000")
	logLine := fmt.Sprintf("[%s][%s] %s", formattedDateTime, chartYaml["name"], cmd)

	// Create log file if not exist
	if _, err := ioutil.ReadFile(logFilePath); err != nil {
		// Ignore if not able to create folder
		_ = os.Mkdir(renderer.debugLogFilePath, 0755)

		f, err := os.Create(logFilePath)
		if err != nil {
			log.Fatalf("Fail to create log file: %v", err)
		}
		f.Close()
	}

	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("Fail to opn debug log file: %v", err)
	}

	// Write log line
	if _, err = f.WriteString(logLine); err != nil {
		log.Fatalf("Fail to write debug log file: %v", err)
	}
}
