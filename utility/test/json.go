//go:build testutils

package test

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

func MustMarshalJson(object interface{}) []byte {
	if object == nil {
		return nil
	}
	data, err := json.Marshal(object)
	Expect(err).ToNot(HaveOccurred())
	return data
}
