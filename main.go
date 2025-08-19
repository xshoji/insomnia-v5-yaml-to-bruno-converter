package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	UsageRequiredPrefix = "\u001B[33m(REQ)\u001B[0m "
	TimeFormat          = "2006-01-02 15:04:05.0000 [MST]"
)

var (
	// Command options ( the -h, --help option is defined by default in the flag package )
	commandDescription         = "This tool converts Insomnia-exported files (v5 YAML) into a Bruno collection files."
	commandOptionFieldWidth    = "12" // recommended width = general: 12, bool only: 5
	optionInsomniaYamlFilePath = flag.String("f" /*  */, "" /*                         */, UsageRequiredPrefix+"Path to Insomnia exported file")
	optionOutputDir            = flag.String("o" /*  */, "" /*                         */, UsageRequiredPrefix+"Output directory")
	optionBrunoCollectionName  = flag.String("n" /*  */, "" /*                         */, UsageRequiredPrefix+"Name of bruno collection")
)

func init() {
	// Format usage
	b := new(bytes.Buffer)
	func() { flag.CommandLine.SetOutput(b); flag.Usage(); flag.CommandLine.SetOutput(os.Stderr) }()
	usage := strings.Replace(strings.Replace(b.String(), ":", " [OPTIONS] [-h, --help]\n\nDescription:\n  "+commandDescription+"\n\nOptions:\n", 1), "Usage of", "Usage:", 1)
	re := regexp.MustCompile(`[^,] +(-\S+)(?: (\S+))?\n*(\s+)(.*)\n`)
	flag.Usage = func() {
		_, _ = fmt.Fprint(flag.CommandLine.Output(), re.ReplaceAllStringFunc(usage, func(m string) string {
			return fmt.Sprintf("  %-"+commandOptionFieldWidth+"s %s\n", re.FindStringSubmatch(m)[1]+" "+strings.TrimSpace(re.FindStringSubmatch(m)[2]), re.FindStringSubmatch(m)[4])
		}))
	}
}

// Build:
// $ GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o /tmp/tool main.go
func main() {

	flag.Parse()
	if *optionInsomniaYamlFilePath == "" || *optionOutputDir == "" || *optionBrunoCollectionName == "" {
		fmt.Printf("\n[ERROR] Missing required option\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Print all options
	fmt.Printf("[ Command options ]\n")
	flag.VisitAll(func(a *flag.Flag) {
		fmt.Printf("  -%-30s %s\n", fmt.Sprintf("%s %v", a.Name, a.Value), strings.Trim(a.Usage, "\n"))
	})
	fmt.Printf("\n\n")

	// Create output directory if it doesn't exist
	if _, err := os.Stat(*optionOutputDir); os.IsNotExist(err) {
		err := os.MkdirAll(*optionOutputDir, 0755)
		handleError(err, "os.MkdirAll(*optionOutputDir, 0755)")
		fmt.Printf("Created output directory: %s\n", *optionOutputDir)
	} else if err != nil {
		handleError(err, "os.Stat(*optionOutputDir)")
	}

	var data map[any]any
	err := yaml.Unmarshal([]byte(ReadAllFileContents(optionInsomniaYamlFilePath)), &data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(data)

	createBrunoJsonFile(*optionOutputDir, *optionBrunoCollectionName)

	createEnvironmentFile(*optionOutputDir, data)

	createCollectionFile(*optionOutputDir, data["collection"].([]any))

}

func createCollectionFile(optionOutputDir string, collectionList []any) {
	for _, item := range collectionList {
		itemMap := item.(map[any]any)
		itemName := itemMap["name"].(string)
		itemId := itemMap["meta"].(map[any]any)["id"].(string)

		if strings.HasPrefix(itemId, "fld_") {
			folder := fmt.Sprintf("%s/%s", optionOutputDir, itemName)
			err := os.MkdirAll(folder, 0755)
			handleError(err, "os.MkdirAll(folder, 0755)")
			fmt.Printf("Created directory for folder: %s\n", itemName)
			createCollectionFile(folder, itemMap["children"].([]any))
		} else if strings.HasPrefix(itemId, "req_") {
			createRequestFile(optionOutputDir, itemMap, itemName, itemId)
		} else {
			handleError(errors.New("Unsupported item type: "+itemId), "createCollectionFile")
		}
	}
}

func createRequestFile(optionOutputDir string, itemMap map[any]any, itemName, itemId string) {
	fileName := strings.ReplaceAll(itemName, "/", "_") + "_" + itemId + ".bru"
	metaData := `meta {
  name: ` + itemName + `
  type: http
  seq: 1
}`

	method := strings.ToLower(itemMap["method"].(string))
	urlAny := itemMap["url"]
	url := ""
	if urlAny != nil {
		url = urlAny.(string)
	}
	body := itemMap["body"]
	bodyType := detectBodyType(body)
	methodData := method + ` {
  url: ` + url + `
  body: ` + bodyType + `
  auth: inherit
}`

	headers := parseHeaders(itemMap)
	headerData := ""
	if len(headers) > 0 {
		var headerVars []string
		for key, value := range headers {
			headerVars = append(headerVars, "  "+key+": "+value)
		}
		headerContent := strings.Join(headerVars, "\n")
		headerData = `headers {
` + headerContent + `
}`
	}

	settingsData := `settings {
  encodeUrl: true
}`
	createAndWriteFile(fmt.Sprintf("%s/%s", optionOutputDir, fileName), strings.Join([]string{
		metaData,
		methodData,
		headerData,
		settingsData,
	}, "\n\n"))
}

func parseHeaders(itemMap map[any]any) map[string]string {
	result := make(map[string]string)
	if itemMap["headers"] == nil {
		return result
	}
	headers := itemMap["headers"].([]any)
	for _, header := range headers {
		if headerMap, ok := header.(map[any]any); ok {
			if key, ok := headerMap["name"].(string); ok {
				if key == "User-Agent" {
					continue
				}
				if value, ok := headerMap["value"].(string); ok {
					result[strings.ToLower(key)] = value
				}
			}
		}
	}
	return result
}

func detectBodyType(body any) string {
	mineType := ""
	if body != nil && body.(map[any]any)["mimeType"] != nil {
		mineType = body.(map[any]any)["mimeType"].(string)
	}

	if strings.HasSuffix(mineType, "application/json") {
		return "json"
	} else if strings.HasSuffix(mineType, "multipart/form-data") {
		return "multipartForm"
	} else if strings.HasSuffix(mineType, "application/x-www-form-urlencoded") {
		return "formUrlEncoded"
	} else {
		return "none"
	}
}

func createEnvironmentFile(optionOutputDir string, data map[any]any) {
	err := os.MkdirAll(fmt.Sprintf("%s/environments", optionOutputDir), 0755)
	handleError(err, "os.MkdirAll(*optionOutputDir, 0755)")

	environments := data["environments"]
	if environments == nil {
		fmt.Println("No environments found in the Insomnia file.")
		return
	}
	subEnvironments := environments.(map[any]any)["subEnvironments"]
	if subEnvironments == nil {
		fmt.Println("No sub-environments found in the Insomnia file.")
		return
	}

	subEnvironmentList := subEnvironments.([]any)
	for _, subEnvironment := range subEnvironmentList {
		subEnvironmentMap := subEnvironment.(map[any]any)
		environmentName := subEnvironmentMap["name"].(string)
		environmentValuesMap := subEnvironmentMap["data"].(map[any]any)
		var envVars []string
		for key, value := range environmentValuesMap {
			envVars = append(envVars, "  "+key.(string)+": "+value.(string))
		}
		envContent := strings.Join(envVars, "\n")
		createAndWriteFile(fmt.Sprintf("%s/environments/%s.bru", optionOutputDir, environmentName), `vars {
`+envContent+`
}`)

	}

}

func createBrunoJsonFile(optionOutputDir, optionBrunoCollectionName string) {
	createAndWriteFile(fmt.Sprintf("%s/bruno.json", optionOutputDir), `{
  "version": "1",
  "name": "`+optionBrunoCollectionName+`",
  "type": "collection",
  "ignore": [
    "node_modules",
    ".git"
  ]
}`)
}

func createAndWriteFile(filePath string, content string) {
	file, err := os.Create(filePath)
	handleError(err, "os.Create(filePath)")
	defer func() { handleError(file.Close(), "file.Close()") }()

	// Write content to file
	_, err = file.WriteString(content)
	handleError(err, "file.WriteString(content)")
}

// =======================================
// File Utils
// =======================================

func ReadAllFileContents(filePath *string) string {
	file, err := os.Open(*filePath)
	handleError(err, "os.Open(*filePath)")
	defer func() { handleError(file.Close(), "file.Close()") }()

	contents, err := io.ReadAll(file)
	handleError(err, "io.ReadAll(file)")
	return string(contents)
}

// =======================================
// Common Utils
// =======================================

func handleError(err error, prefixErrMessage string) {
	if err != nil {
		fmt.Printf("%s [ERROR %s]: %v\n", time.Now().Format(TimeFormat), prefixErrMessage, err)
	}
}
