// Copyright 2018 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const codeTemplate string = `
/* Code generated by openapi-gen/main.go. DO NOT EDIT. */

namespace Nakama
{
    using System;
    using System.Collections.Generic;
    using System.Threading.Tasks;

    {{- range $defname, $definition := .Definitions }}
    {{- $classname := $defname | title }}

    /// <summary>
    /// {{ $definition.Description | stripNewlines }}
    /// </summary>
    public interface I{{ $classname }}
    {
        {{- range $propname, $property := $definition.Properties }}
        {{- $fieldname := $propname | snakeCaseToPascalCase }}

        /// <summary>
        /// {{ $property.Description }}
        /// </summary>
        {{- if eq $property.Type "integer"}}
        int {{ $fieldname }} { get; }
        {{- else if eq $property.Type "boolean" }}
        bool {{ $fieldname }} { get; }
        {{- else if eq $property.Type "string"}}
        string {{ $fieldname }} { get; }
        {{- else if eq $property.Type "array"}}
          {{- if eq $property.Items.Type "string"}}
        List<string> {{ $fieldname }} { get; }
          {{- else if eq $property.Items.Type "integer"}}
        List<int> {{ $fieldname }} { get; }
          {{- else if eq $property.Items.Type "boolean"}}
        List<bool> {{ $fieldname }} { get; }
          {{- else}}
        List<I{{ $property.Items.Ref | cleanRef }}> {{ $fieldname }} { get; }
          {{- end }}
        {{- else }}
        I{{ $property.Ref | cleanRef }} {{ $fieldname }} { get; }
        {{- end }}
        {{- end }}
    }

    /// <inheritdoc />
    internal class {{ $classname }} : I{{ $classname }}
    {
        {{- range $propname, $property := $definition.Properties }}
        {{- $fieldname := $propname | snakeCaseToPascalCase }}

        /// <inheritdoc />
        {{- if eq $property.Type "integer"}}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public int {{ $fieldname }} { get; set; }
        {{- else if eq $property.Type "boolean" }}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public bool {{ $fieldname }} { get; set; }
        {{- else if eq $property.Type "string"}}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public string {{ $fieldname }} { get; set; }
        {{- else if eq $property.Type "array"}}
          {{- if eq $property.Items.Type "string"}}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public List<string> {{ $fieldname }} { get; set; }
          {{- else if eq $property.Items.Type "integer"}}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public List<int> {{ $fieldname }} { get; set; }
          {{- else if eq $property.Items.Type "boolean"}}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public List<bool> {{ $fieldname }} { get; set; }
          {{- else}}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public List<I{{ $property.Items.Ref | cleanRef }}> {{ $fieldname }} { get; set; }
          {{- end }}
        {{- else }}
        [TinyJson.JsonProperty("{{ $propname }}")]
        public I{{ $property.Ref | cleanRef }} {{ $fieldname }} { get; set; }
        {{- end }}
        {{- end }}

        public override string ToString()
        {
            var output = "";
            {{- range $fieldname, $property := $definition.Properties }}
            output += string.Concat(output, "{{ $fieldname | snakeCaseToPascalCase }}: {", {{ $fieldname | snakeCaseToPascalCase }}, "}, ");
            {{- end}}
            return output;
        }
    }
    {{- end }}
/*
    /// <summary>
    /// The low level client for the Nakama API.
    /// </summary>
    internal class ApiClient
    {
        {{- range $url, $path := .Paths }}
        {{- range $method, $operation := $path}}

        /// <summary>
        /// {{ $operation.Summary | stripNewlines }}
        /// </summary>
        public Task<> {{ $operation.OperationId | snakeCaseToPascalCase }}Async()
        {
        }
        {{- end }}
        {{- end }}
    }
*/
}
`

func convertRefToClassName(input string) (className string) {
	cleanRef := strings.TrimLeft(input, "#/definitions/")
	className = strings.Title(cleanRef)
	return
}

func snakeCaseToPascalCase(input string) (output string) {
	isToUpper := false
	for k, v := range input {
		if k == 0 {
			output = strings.ToUpper(string(input[0]))
		} else {
			if isToUpper {
				output += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					output += string(v)
				}
			}
		}
	}
	return
}

func stripNewlines(input string) (output string) {
	output = strings.Replace(input, "\n", " ", -1)
	return
}

func main() {
	// Argument flags
	var output = flag.String("output", "", "The output for generated code.")
	flag.Parse()

	inputs := flag.Args()
	if len(inputs) < 1 {
		fmt.Printf("No input file found: %s\n\n", inputs)
		fmt.Println("openapi-gen [flags] inputs...")
		flag.PrintDefaults()
		return
	}

	input := inputs[0]
	content, err := ioutil.ReadFile(input)
	if err != nil {
		fmt.Printf("Unable to read file: %s\n", err)
		return
	}

	var schema struct {
		Paths map[string]map[string]struct {
			Summary     string
			OperationId string
			Responses   struct {
				Ok struct {
					Schema struct {
						Ref string `json:"$ref"`
					}
				} `json:"200"`
			}
			Parameters []struct {
				Name     string
				In       string
				Required bool
				Type     string   // used with primitives
				Items    struct { // used with type "array"
					Type string
				}
				Schema struct { // used with http body
					Type string
					Ref  string `json:"$ref"`
				}
			}
		}
		Definitions map[string]struct {
			Properties map[string]struct {
				Type  string
				Ref   string   `json:"$ref"` // used with object
				Items struct { // used with type "array"
					Type string
					Ref  string `json:"$ref"`
				}
				Format      string // used with type "boolean"
				Description string
			}
			Description string
		}
	}

	if err := json.Unmarshal(content, &schema); err != nil {
		fmt.Printf("Unable to decode input %s : %s\n", input, err)
		return
	}

	fmap := template.FuncMap{
		"cleanRef":              convertRefToClassName,
		"snakeCaseToPascalCase": snakeCaseToPascalCase,
		"stripNewlines":         stripNewlines,
		"title":                 strings.Title,
		"uppercase":             strings.ToUpper,
	}
	tmpl, err := template.New(input).Funcs(fmap).Parse(codeTemplate)
	if err != nil {
		fmt.Printf("Template parse error: %s\n", err)
		return
	}

	if len(*output) < 1 {
		tmpl.Execute(os.Stdout, schema)
		return
	}

	f, err := os.Create(*output)
	if err != nil {
		fmt.Printf("Unable to create file: %s\n", err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	tmpl.Execute(writer, schema)
	writer.Flush()
}
