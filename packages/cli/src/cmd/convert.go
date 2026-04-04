package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var inputFile string
var outputFile string

// ─── Types ────────────────────────────────────────────────────────────────────

type JSON map[string]interface{}

// ─── YAML Normalizer (FIX) ────────────────────────────────────────────────────

// Converts map[interface{}]interface{} → map[string]interface{}
func normalizeYAML(i interface{}) interface{} {
	switch v := i.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, val := range v {
			m[fmt.Sprintf("%v", k)] = normalizeYAML(val)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = normalizeYAML(val)
		}
	}
	return i
}

// ─── Parser ───────────────────────────────────────────────────────────────────

func parseOpenAPI(data []byte) (JSON, error) {
	var spec JSON

	// Try JSON first
	if err := json.Unmarshal(data, &spec); err == nil {
		return spec, nil
	}

	// YAML fallback
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	normalized := normalizeYAML(raw)

	result, ok := normalized.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse OpenAPI spec")
	}

	return result, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func getMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func getSlice(v interface{}) []interface{} {
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ─── Schema → Example ─────────────────────────────────────────────────────────

func schemaToExample(schema map[string]interface{}) interface{} {
	if schema == nil {
		return nil
	}

	if ex, ok := schema["example"]; ok {
		return ex
	}

	if def, ok := schema["default"]; ok {
		return def
	}

	typ := getString(schema["type"])

	switch typ {
	case "string":
		if enum, ok := schema["enum"].([]interface{}); ok && len(enum) > 0 {
			return enum[0]
		}
		return "string"

	case "integer", "number":
		return 0

	case "boolean":
		return true

	case "array":
		items := getMap(schema["items"])
		return []interface{}{schemaToExample(items)}

	case "object":
		props := getMap(schema["properties"])
		obj := map[string]interface{}{}
		for k, v := range props {
			obj[k] = schemaToExample(getMap(v))
		}
		return obj
	}

	return nil
}

// ─── Converter ────────────────────────────────────────────────────────────────

func convertToPostman(spec JSON) (JSON, error) {
	info := JSON{
		"name":        "Imported Collection",
		"_postman_id": "auto-generated",
		"description": "",
		"schema":      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
	}

	if i := getMap(spec["info"]); i != nil {
		if t := getString(i["title"]); t != "" {
			info["name"] = t
		}
		if d := getString(i["description"]); d != "" {
			info["description"] = d
		}
	}

	baseURL := ""
	if servers := getSlice(spec["servers"]); len(servers) > 0 {
		if s := getMap(servers[0]); s != nil {
			baseURL = getString(s["url"])
		}
	}

	tagMap := map[string][]JSON{}

	paths := getMap(spec["paths"])
	if paths == nil {
		return nil, fmt.Errorf("invalid paths (check YAML structure)")
	}

	for path, methodsRaw := range paths {
		methods := getMap(methodsRaw)

		for method, opRaw := range methods {
			op := getMap(opRaw)

			tag := "default"
			if tags := getSlice(op["tags"]); len(tags) > 0 {
				tag = getString(tags[0])
			}

			name := strings.ToUpper(method) + " " + path
			if s := getString(op["summary"]); s != "" {
				name = s
			}

			var query []JSON
			var pathVars []JSON
			var headers []JSON

			for _, pRaw := range getSlice(op["parameters"]) {
				p := getMap(pRaw)
				in := getString(p["in"])

				param := JSON{
					"key":         getString(p["name"]),
					"value":       "",
					"description": getString(p["description"]),
				}

				if ex := p["example"]; ex != nil {
					param["value"] = fmt.Sprintf("%v", ex)
				}

				switch in {
				case "query":
					query = append(query, param)
				case "path":
					pathVars = append(pathVars, param)
				case "header":
					headers = append(headers, param)
				}
			}

			// Body
			var body interface{}
			content := getMap(getMap(op["requestBody"])["content"])

			if mt := getMap(content["application/json"]); mt != nil {
				ex := mt["example"]

				if ex == nil {
					if examples := getMap(mt["examples"]); examples != nil {
						for _, v := range examples {
							ex = getMap(v)["value"]
							break
						}
					}
				}

				if ex == nil {
					ex = schemaToExample(getMap(mt["schema"]))
				}

				raw, _ := json.MarshalIndent(ex, "", "  ")

				body = JSON{
					"mode": "raw",
					"raw":  string(raw),
					"options": JSON{
						"raw": JSON{"language": "json"},
					},
				}

				headers = append(headers, JSON{
					"key":   "Content-Type",
					"value": "application/json",
				})
			}

			rawURL := baseURL + path

			req := JSON{
				"method": strings.ToUpper(method),
				"header": headers,
				"url": JSON{
					"raw":      rawURL,
					"path":     strings.Split(strings.Trim(path, "/"), "/"),
					"query":    query,
					"variable": pathVars,
				},
				"description": getString(op["description"]),
			}

			if body != nil {
				req["body"] = body
			}

			item := JSON{
				"name":     name,
				"request":  req,
				"response": []interface{}{},
			}

			tagMap[tag] = append(tagMap[tag], item)
		}
	}

	var folders []JSON
	for tag, items := range tagMap {
		folders = append(folders, JSON{
			"name": tag,
			"item": items,
		})
	}

	return JSON{
		"info": info,
		"item": folders,
		"variable": []JSON{
			{
				"key":   "baseUrl",
				"value": baseURL,
				"type":  "string",
			},
		},
	}, nil
}

// ─── Cobra Command ────────────────────────────────────────────────────────────

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert OpenAPI to Postman collection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if inputFile == "" {
			return fmt.Errorf("input file required (-i)")
		}

		data, err := os.ReadFile(inputFile)
		if err != nil {
			return err
		}

		spec, err := parseOpenAPI(data)
		if err != nil {
			return err
		}

		postman, err := convertToPostman(spec)
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(postman, "", "  ")
		if err != nil {
			return err
		}

		if outputFile == "" {
			fmt.Println(string(out))
			return nil
		}

		return os.WriteFile(outputFile, out, 0644)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&inputFile, "input", "i", "", "OpenAPI file (json/yaml)")
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output Postman file")
}
