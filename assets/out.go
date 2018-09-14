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

	method, data, err, empty := common.ValidateAndBuildPostBody(input)
	common.HandleFatalError(err, "Error while validating input")

	if empty {
		fmt.Println("skipping sending a message because there is no text")
	}
	var response = common.SlackResponse{}
	response, err = common.PostToSlack(method, data)
	common.HandleFatalError(err, "Error while posting to slack")
	_ = response

	// Change this to the ID of the returned resource
	fmt.Println(`{"version":{"ref":"none"}}`)
}
