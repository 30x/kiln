package shipyard_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestShipyard(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shipyard Suite")
}
