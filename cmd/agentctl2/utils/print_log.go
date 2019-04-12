package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"os"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

type logType struct {
	Logger string `json:"logger,omitempty"`
	Level  string `json:"level,omitempty"`
}

type LogList []logType

func ConvertToLogList(log string) LogList {
	data := make(LogList, 0)
	err := json.Unmarshal([]byte(log), &data)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed conver string to json - "+err.Error()))
	}

	return data
}

func (ll LogList) PrintLogList() (*bytes.Buffer, error) {

	logLevel := createLogTypeTemplate()

	templates := []*template.Template{}

	templates = append(templates, logLevel)

	return ll.textRenderer(templates)
}

func createLogTypeTemplate() *template.Template {
	Template, err := template.New("log").Parse(
		"{{with .Logger}}\nLogger: {{.}}{{end}}" +
			"{{with .Level}}\nLevel: {{.}}{{end}}")

	if err != nil {
		panic(err)
	}

	return Template
}

func (ll LogList) textRenderer(templates []*template.Template) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	buffer.WriteTo(os.Stdout)
	for _, value := range ll {
		var wasError error
		for _, templateVal := range templates {
			wasError = templateVal.Execute(buffer, value)
			if nil != wasError {
				return nil, wasError
			}
		}
	}

	return buffer, nil
}
