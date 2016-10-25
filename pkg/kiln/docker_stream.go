package kiln

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

//StreamParser a stream parser for parsing different docker response
type StreamParser interface {
	//Return the channel to receive parsed events from
	Channel() chan (string)
	//ParseStream A stream parser for different stream responses from docker.  Should return a string for each message we need to return
	Parse()
}

//NewBuildStreamParser create a new build stream parser
func NewBuildStreamParser(stream io.Reader) StreamParser {

	outputChannel := make(chan string, getBufferSize())

	return &BuildStreamParser{
		stream:        stream,
		outputChannel: outputChannel,
	}
}

//BuildStreamParser a build stream parser
type BuildStreamParser struct {
	stream        io.Reader
	outputChannel chan (string)
}

//Parse parse the stream and emit the values into the channel
func (parser *BuildStreamParser) Parse() {

	//create a scanner to scan lines, since this isn't actually well formed json
	scanner := bufio.NewScanner(parser.stream)

	scanner.Split(bufio.ScanLines)

	LogInfo.Printf("Starting to parse build stream")

	for scanner.Scan() {

		line := scanner.Text()

		//parse the json
		var parsed streamStatusMessage

		err := json.Unmarshal([]byte(line), &parsed)

		//can't parse it, discard it. Not a line we care about
		if err != nil || parsed.Stream == "" {
			continue
		}

		parser.outputChannel <- parsed.Stream
	}

	LogInfo.Printf("Completed parsing build stream")

	close(parser.outputChannel)
}

//Channel return the channel that was allocated
func (parser *BuildStreamParser) Channel() chan (string) {
	return parser.outputChannel
}

//NewPushStreamParser create a new push stream parser
func NewPushStreamParser(stream io.Reader) StreamParser {

	outputChannel := make(chan string, getBufferSize())

	return &PushStreamParser{
		stream:        stream,
		outputChannel: outputChannel,
	}
}

//PushStreamParser a build stream parser
type PushStreamParser struct {
	stream        io.Reader
	outputChannel chan (string)
}

//Parse parse the stream and emit the values into the channel
func (parser *PushStreamParser) Parse() {

	//create a scanner to scan lines, since this isn't actually well formed json
	scanner := bufio.NewScanner(parser.stream)

	scanner.Split(bufio.ScanLines)

	LogInfo.Printf("Starting to parse push stream")

	for scanner.Scan() {

		line := scanner.Text()

		//parse the json
		var pushStatusMessage pushStatusMessage

		err := json.Unmarshal([]byte(line), &pushStatusMessage)

		//can't parse it, discard it. Not a line we care about
		if err != nil {
			continue
		}

		if pushStatusMessage.ProgressDetail.Current == 0 && pushStatusMessage.ProgressDetail.Total == 0 {
			continue;
		} else {
			text := fmt.Sprintf("current uploaded:%d, total size:%d\n", pushStatusMessage.ProgressDetail.Current, pushStatusMessage.ProgressDetail.Total)

			parser.outputChannel <- text
		}
	}

	LogInfo.Printf("Completed parsing push stream")

	close(parser.outputChannel)
}

//Channel return the channel that was allocated
func (parser *PushStreamParser) Channel() chan (string) {
	return parser.outputChannel
}

func getBufferSize() int {

	sizeString := os.Getenv("BUFFER_LINES")

	//default is 100
	if sizeString == "" {
		return 100
	}

	intVal, err := strconv.Atoi(sizeString)

	if err != nil {
		return 100
	}

	return intVal

}

// {"status":"Pushing","progressDetail":{"current":512,"total":1598},"progress":"[================\u003e                                  ]    512 B/1.598 kB","id":"715751c25079"}

//this is the only status message we care about

type pushStatusMessage struct {
	ID             string `json:"id"`       // 715751c25079
	Progress       string `json:"progress"` // [================>                                  ]    512 B/1.598 kB
	ProgressDetail struct {
		Current int64 `json:"current"` // 512
		Total   int64 `json:"total"`   // 1598
	} `json:"progressDetail"`
	Status string `json:"status"` // Pushing
}

type streamStatusMessage struct {
	Stream string `json:"stream"`
}
