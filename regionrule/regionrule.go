package regionrule

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
)

type Rule struct {
	Name       string `yaml:"name"`
	ISD        int    `yaml:"ISD"`
	Preference string `yaml:"Preference"`
}

type AppConfig struct {
	Apps []struct {
		Name  string `yaml:"name"`
		Rules []Rule `yaml:"rules"`
	} `yaml:"apps"`
}

func GetPreferences() ([]string, error) {
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
	fmt.Printf("Parsed YAML Configuration:\n")
	var preferences []string
	for _, app := range config.Apps {
		for _, rule := range app.Rules {
			preferences = append(preferences, rule.Preference)
		}
	}

	return preferences, nil
}
