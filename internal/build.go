package internal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	defaultRepoConfigPath               = "/helm-working-dir/"
	defaultHelmRegistrySecretConfigPath = "/helm-working-dir/plugin-repositories/repositories.yaml"
	authHelmRegistry                    = []string{"https://gitlab.int.hextech.io"}
)

type HelmRepositoryConfig struct {
	ApiVersion   string       `default:"" yaml:"apiVersion"`
	Generated    string       `default:"0001-01-01T00:00:00Z" yaml:"generated"`
	Repositories []Repository `yaml:"repositories"`
}

type Repository struct {
	CaFile                string `default:"" yaml:"caFile"`
	CertFile              string `default:"" yaml:"certFile"`
	InsecureSkipTlsVerify bool   `default:false yaml:"insecure_skip_tls_verify"`
	KeyFile               string `default:"" yaml:"keyFile"`
	Name                  string `default:"" yaml:"name"`
	PassCredentialsAll    bool   `default:false yaml:"pass_credentials_all"`
	Username              string `default:"" yaml:"username"`
	Password              string `default:"" yaml:"password"`
	Url                   string `default:"" yaml:"url"`
}

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (builder *Builder) Build(helmChartPath string, repoConfigPath string, helmRegistrySecretConfigPath string) {
	if len(helmChartPath) <= 0 {
		helmChartPath = defaultHelmChartPath
	}
	if len(repoConfigPath) <= 0 {
		repoConfigPath = defaultRepoConfigPath
	}
	if len(helmRegistrySecretConfigPath) <= 0 {
		helmRegistrySecretConfigPath = defaultHelmRegistrySecretConfigPath
	}

	os.Chdir(helmChartPath)
	useExternalHelmChartPathIfSet()
	chartYaml := ReadChartYaml()

	// Skip if chart doesn't have dependency
	dependencies := chartYaml["dependencies"]
	if dependencies == nil || len(dependencies.([]interface{})) <= 0 {
		log.Println("Not dependency found.")
		return
	}

	// Use app name as config file name
	repositoryConfigName := repoConfigPath + chartYaml["name"].(string) + ".yaml"
	log.Printf("repositoryConfigName: %s\n", repositoryConfigName)

	builder.generateRepositoryConfig(repositoryConfigName, chartYaml, helmRegistrySecretConfigPath)
	builder.executeHelmDependencyBuild(repositoryConfigName)
}

func (builder *Builder) generateRepositoryConfig(repositoryConfigName string, chartYaml map[string]interface{}, helmRegistrySecretConfigPath string) {
	repos := []Repository{}
	// Read dependencies from Chart.yaml, and generate repositories.yaml from it
	for _, dep := range chartYaml["dependencies"].([]interface{}) {
		d := dep.(map[interface{}]interface{})
		repositoryUrl := d["repository"].(string)
		name := d["name"].(string)
		username := ""
		password := ""
		for _, authReg := range authHelmRegistry {
			if strings.HasPrefix(repositoryUrl, authReg) {
				// Read username password from /helm-working-dir/plugin-repositories/repositories.yaml
				u, p := builder.readRepositoryConfig(repositoryUrl, helmRegistrySecretConfigPath)
				username = u
				password = p
				break
			}
		}

		repos = append(repos, Repository{
			Name:     name,
			Url:      repositoryUrl,
			Username: username,
			Password: password,
		})
	}

	repoConfig := HelmRepositoryConfig{
		Generated:    "0001-01-01T00:00:00Z",
		Repositories: repos,
	}

	yamlConfig, err := yaml.Marshal(repoConfig)
	if err != nil {
		log.Fatalf("Marshal helm repository yaml error: %v", err)
	}
	fmt.Println(string(yamlConfig))

	err = os.WriteFile(repositoryConfigName, []byte(yamlConfig), 0777)
	if err != nil {
		log.Fatalf("Write helm repository yaml error: %v", err)
	}
}

func (builder *Builder) readRepositoryConfig(repositoryUrl string, helmRegistrySecretConfigPath string) (string, string) {
	repo := HelmRepositoryConfig{}

	// Read helm repository config created by Terraform
	bs, err := ioutil.ReadFile(helmRegistrySecretConfigPath)
	if err != nil {
		log.Fatalf("Error reading repositories.yaml: %v", err)
	}

	if err := yaml.Unmarshal(bs, &repo); err != nil {
		log.Fatalf("Error unmarshal helmRegistrySecretConfigPath: %v", err)
	}

	// Return the username password if url matches
	for _, r := range repo.Repositories {
		if r.Url == repositoryUrl {
			return r.Username, r.Password
		}
	}
	return "", ""
}

func (builder *Builder) executeHelmDependencyBuild(repositoryConfigName string) {
	command := "helm"
	args := []string{"dependency", "build", "--repository-config", repositoryConfigName}
	cmd := exec.Command(command, args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Exec helm dependency build error: %s\n%s", err, stderr.String())
	}
	log.Println(out.String())
}
