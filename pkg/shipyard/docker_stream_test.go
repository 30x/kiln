package shipyard_test

import (
	"strings"

	. "github.com/30x/shipyard/pkg/shipyard"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// "github.com/docker/engine-api/types"
)

var _ = Describe("Docker Stream", func() {

	It("Test Build Parsing", func() {

		testStream :=
			`{"stream":"Step 1 : FROM mhart/alpine-node:4\n"}
{"status":"Pulling from mhart/alpine-node","id":"4"}
{"status":"Pulling fs layer","progressDetail":{},"id":"d0ca440e8637"}
{"status":"Pulling fs layer","progressDetail":{},"id":"928fef61ff1e"}
{"status":"Downloading","progressDetail":{"current":32264,"total":2320212},"progress":"[\u003e                                                  ] 32.26 kB/2.32 MB","id":"d0ca440e8637"}
{"status":"Downloading","progressDetail":{"current":64465,"total":2320212},"progress":"[=\u003e                                                 ] 64.47 kB/2.32 MB","id":"d0ca440e8637"}
{"status":"Downloading","progressDetail":{"current":97233,"total":2320212},"progress":"[==\u003e                                                ] 97.23 kB/2.32 MB","id":"d0ca440e8637"}
{"status":"Downloading","progressDetail":{"current":130001,"total":2320212},"progress":"[==\u003e           
{"stream":" ---\u003e 8a54b3f8ac77\n"}
{"stream":"Step 2 : ADD . .\n"}`

		streamParser := NewBuildStreamParser(strings.NewReader(testStream))

		//start parsing
		go streamParser.Parse()

		channel := streamParser.Channel()

		value := <-channel

		Expect(value).Should(Equal("Step 1 : FROM mhart/alpine-node:4\n"))

		value = <-channel

		Expect(value).Should(Equal(" ---\u003e 8a54b3f8ac77\n"))

		value = <-channel

		Expect(value).Should(Equal("Step 2 : ADD . .\n"))

		//should be closed
		value, ok := <-channel

		Expect(value).Should(BeEmpty())

		Expect(ok).Should(BeFalse())
	})

	It("Test Push Parsing", func() {

		testStream :=
			`{"status":"Pushing","progressDetail":{"current":512,"total":1598},"progress":"[================\u003e                                  ]    512 B/1.598 kB","id":"715751c25079"}
{"status":"Pushing","progressDetail":{"current":751,"total":1598},"progress":"[=======================\u003e                           ]    751 B/1.598 kB","id":"715751c25079"}
{"status":"Pushing","progressDetail":{"current":1536,"total":1598},"progress":"[================================================\u003e  ] 1.536 kB/1.598 kB","id":"715751c25079"}
{"status":"Pushing","progressDetail":{"current":2589,"total":1598},"progress":"[==================================================\u003e] 2.589 kB","id":"715751c25079"}`
		streamParser := NewPushStreamParser(strings.NewReader(testStream))

		//start parsing
		go streamParser.Parse()

		channel := streamParser.Channel()

		value := <-channel

		Expect(value).Should(Equal("current uploaded:512, total size:1598"))

		value = <-channel

		Expect(value).Should(Equal("current uploaded:751, total size:1598"))

		value = <-channel

		Expect(value).Should(Equal("current uploaded:1536, total size:1598"))

		value = <-channel

		Expect(value).Should(Equal("current uploaded:2589, total size:1598"))

		//should be closed
		value, ok := <-channel

		Expect(value).Should(BeEmpty())

		Expect(ok).Should(BeFalse())
	})
})
