package internal

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func ReadChartYaml() map[string]interface{} {
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
