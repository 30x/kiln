package kiln_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKiln(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kiln Suite")
}
