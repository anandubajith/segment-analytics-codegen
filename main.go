package main // TODO: figure out how to keep main in this

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type PropertySpec struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type EventSpec struct {
	Name       string         `yaml:"name"`
	Properties []PropertySpec `yaml:"properties"`
}

type Spec struct {
	Events []EventSpec `yaml:"events"`
}

const codeTemplate = `package analytics

import (
	"fmt"
	"github.com/segmentio/analytics-go"
)


type AnalyticsClient struct {
	client analytics.Client
}

func NewAnalyticsClient(writeKey string) *AnalyticsClient {
	client := analytics.New(writeKey)
	return &AnalyticsClient{client}
}

func (c *AnalyticsClient) Close() {
	c.client.Close()
}

{{range .Events}}
type {{.Name | Title}}EventProps struct {
    {{range .Properties}}{{.Name | Title}} {{.Type | GoType}}
    {{end}}
}

func (c *AnalyticsClient) Track{{.Name | Title}}Event(properties {{.Name | Title}}EventProps) {
    fmt.Printf("Tracking {{.Name}} event: %+v\n", properties)

		 c.client.Enqueue(analytics.Track{
        Event:  "{{.Name}}",
				Properties: analytics.NewProperties(){{range .Properties}}.Set("{{.Name }}", properties.{{.Name | Title}}){{end}},
    })
}
{{end}}`

func goType(yamlType string) string {
	switch yamlType {
	case "string":
		return "string"
	case "int":
		return "int"
	case "bool":
		return "bool"
	default:
		return "interface{}"
	}
}

func main() {
	// Read the input YAML file
	if len(os.Args) < 3 {
		panic("Usage: main.go <input.yaml> <output.go>")
	}
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		panic(err)
	}

	// Create the template
	tmpl, err := template.New("code").Funcs(template.FuncMap{
		"Title":  strings.Title,
		"GoType": goType,
	}).Parse(string(codeTemplate))

	if err != nil {
		panic(err)
	}

	// Generate the code
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		panic(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Write to the output file
	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		panic(err)
	}

	// TODO: format the output file

	fmt.Printf("Code generated successfully in %s\n", outputFile)
}
