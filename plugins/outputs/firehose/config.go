package firehose

import (
	"bytes"
	"os"
	"regexp"
	"strings"
)

func escapeEnv(value string) string {
	return strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	).Replace(value)
}

func parseConfig(contents []byte) []byte {
	envVarRe := regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)

	parameters := envVarRe.FindAllSubmatch(contents, -1)
	for _, parameter := range parameters {
		if len(parameter) != 3 {
			continue
		}

		var envVar []byte
		if parameter[1] != nil {
			envVar = parameter[1]
		} else if parameter[2] != nil {
			envVar = parameter[2]
		} else {
			continue
		}

		envVal, ok := os.LookupEnv(strings.TrimPrefix(string(envVar), "$"))
		if ok {
			envVal = escapeEnv(envVal)
			contents = bytes.Replace(contents, parameter[0], []byte(envVal), 1)
		}
	}

	return contents
}
