package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKiln(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kiln Server Suite")
}
