package regionrule

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

type Rule struct {
	Name          string `yaml:"name"`
	Permitted_ISD []int  `yaml:"Permitted_ISD"`
	Preference    string `yaml:"Preference"`
}

type AppConfig struct {
	Apps []struct {
		Name  string `yaml:"name"`
		Rules []Rule `yaml:"rules"`
	} `yaml:"apps"`
}

func GetRules() ([]Rule, error) {
	fileURL := "https://raw.githubusercontent.com/alicecuii/HelloSCIONWorld/main/configfiles/app.yml"
	// Make an HTTP GET request to the YAML file
	response, err := http.Get(fileURL)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer response.Body.Close()

	yamlData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Unmarshal the YAML data into the AppConfig struct
	var config AppConfig
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}
	// Read the YAML content from the response

	// Print the parsed configuration to the screen
	// fmt.Printf("Parsed YAML Configuration:\n")
	var rules []Rule
	for _, app := range config.Apps {
		for _, rule := range app.Rules {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}
