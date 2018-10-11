package main

import (
	"fmt"
	"os"

	"./common"
)

func main() {
	os.Chdir(os.Args[1])

	input, err := common.GetInput()
	common.HandleFatalError(err, "Error getting concourse input")

	method, data, empty, err := common.ValidateAndBuildPostBody(input)
	common.HandleFatalError(err, "Error while validating input")
	if !empty {
		_, err = common.PostToSlack(method, data, input.Params.FallbackChannel, true)
		common.HandleFatalError(err, "Error while posting to slack")
	}

	// Change this to the ID of the returned resource
	fmt.Println(`{"version":{"ref":"none"}}`)
}
