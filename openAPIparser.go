package OAPIClientGenerator

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type OAPIEndpoint struct {
	Summary     string                  `yaml:"summary"`
	OperationID string                  `yaml:"operationId"`
	Tags        []string                `yaml:"tags"`
	Parameters  []OAPIParameter         `yaml:"parameters"`
	Responses   map[string]OAPIResponse `yaml:"responses"`
}

type OAPIParameter struct {
	Name        string `yaml:"name"`
	In          string `yaml:"in"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Type        string `yaml:"type"`
	Format      string `yaml:"format"`
	Ref         string `yaml:"$ref"`
}

type OAPIResponse struct {
	Description string `yaml:"description"`
	Headers     map[string]struct {
		Type        string `yaml:"type"`
		Description string `yaml:"description"`
	} `yaml:"headers"`
	Schema struct {
		Ref string `yaml:"$ref"`
	} `yaml:"schema"`
}

type OAPIDefinition struct {
	Type       string   `yaml:"type"`
	Required   []string `yaml:"required"`
	Properties map[string]struct {
		Type   string `yaml:"type"`
		Format string `yaml:"format"`
	} `yaml:"properties"`
	Items struct {
		Ref string `yaml:"$ref"`
	} `yaml:"items"`
}

func (o OAPIDefinition) isArray() bool {
	return o.Type == "array"
}

func (o OAPIDefinition) isStructure() bool {
	return o.Type == "object"
}

const structBase = `
USTRUCT(Blueprintable)
struct %s {
	GENERATED_BODY()
`
const propertyBase = `	UPROPERTY(EditAnywhere, BlueprintReadWrite)
	%s
`
const structEnd = `
};
`

func withUEStructPrefix(name string) string {
	return "F" + name
}

func (o OAPIDefinition) generateCppStructure(name string) string {
	if o.Type == "object" {
		structure := fmt.Sprintf(structBase, withUEStructPrefix(name))
		for propertyName, propertyType := range o.Properties {
			cppType := getCppType(propertyType.Type, propertyType.Format)
			structure += fmt.Sprintf(propertyBase, cppType+" "+strings.Title(propertyName)+";")
		}
		structure += structEnd
		return structure
	}
	return ""
}

func getCppType(theType, format string) string {
	switch theType {
	case "integer":
		switch format {
		case "int32":
			return "int32"
		default:
			return "int"
		}
	case "string":
		return "FString"
	}
	return ""
}

type OAPIMethod map[string]OAPIEndpoint

type OAPI struct {
	Swagger string `yaml:"swagger"`
	Info    struct {
		Version string `yaml:"version"`
		Title   string `yaml:"title"`
		License struct {
			Name string `yaml:"name"`
		} `yaml:"license"`
	} `yaml:"info"`
	Host        string                    `yaml:"host"`
	BasePath    string                    `yaml:"basePath"`
	Schemes     []string                  `yaml:"schemes"`
	Consumes    []string                  `yaml:"consumes"`
	Produces    []string                  `yaml:"produces"`
	Paths       map[string]OAPIMethod     `yaml:"paths"`
	Definitions map[string]OAPIDefinition `yaml:"definitions"`
}

func (o *OAPI) Parse(filename string) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("%s read error #%v ", filename, err)
	}
	err = yaml.Unmarshal(yamlFile, o)
	if err != nil {
		log.Fatalf("%s unmarshal: %v", filename, err)
	}
	return nil
}
