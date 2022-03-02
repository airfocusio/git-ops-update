package internal

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func readYaml(bytes []byte, v interface{}) error {
	return yaml.Unmarshal(bytes, v)
}

func writeYaml(v interface{}) ([]byte, error) {
	var bs bytes.Buffer
	enc := yaml.NewEncoder(&bs)
	enc.SetIndent(2)
	err := enc.Encode(v)
	return bs.Bytes(), err
}
