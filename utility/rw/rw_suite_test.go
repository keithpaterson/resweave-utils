package rw_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRw(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rw Suite")
}
