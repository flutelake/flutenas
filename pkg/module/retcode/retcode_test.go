package retcode

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type code struct {
	Name    string `json:"name"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Test_generateRetCodes(t *testing.T) {
	currentDir, _ := os.Getwd()
	targetDir := filepath.Join(currentDir, "code")
	entries, _ := os.ReadDir(targetDir)

	generatedFile := filepath.Join(currentDir, "retcode_generate.go")
	f, err := os.OpenFile(generatedFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}

	f.WriteString("// Code generated by fluteNAS. DO NOT EDIT.\n\n")

	f.WriteString("package retcode\n\n")

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		bs, err := os.ReadFile(filepath.Join(targetDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		codes := []code{}
		yaml.Unmarshal(bs, &codes)

		for _, c := range codes {
			f.WriteString(fmt.Sprintf(`var Status%s = func(data any) *RetCode { return &RetCode{Code: %d, Message: "%s", Data: data}}`, c.Name, c.Code, c.Message))
			f.WriteString("\n")
		}
	}
}
