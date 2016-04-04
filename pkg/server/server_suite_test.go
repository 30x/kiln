package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestShipyard(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shipyard Server Suite")
}
