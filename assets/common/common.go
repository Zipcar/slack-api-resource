package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Reusable constants
const (
	SlackAPIURL           = "https://slack.com"
	MethodFilesUpload     = "files.upload"
	MethodChatPostmessage = "chat.postMessage"
)

var slackMethodPaths = map[string]string{
	MethodFilesUpload:     "api/files.upload",
	MethodChatPostmessage: "api/chat.postMessage",
}

// GetInput ... Retrieves input from stdio
func GetInput() (ConcourseInput, error) {
	input := ConcourseInput{}

	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &input)
		if err != nil {
			return input, err
		}

		return input, nil
	}

	return input, errors.New("no input received")
}

// HandleFatalError ... Exits with a non-zero status and prints error to stderr if error is not nil, does nothing otherwise
func HandleFatalError(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err.Error()))
	os.Exit(1)
}

// HandleNonFatalError ... Prints error to stderr if error is not nil, does nothing otherwise
func HandleNonFatalError(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err.Error()))
}

// ValidateAndBuildPostBody ... Directs requests to appropriate delegates depending on if we deal with a file upload or a chat message
func ValidateAndBuildPostBody(input ConcourseInput) (method string, values url.Values, emtpy bool, err error) {
	var exists = false
	if method, exists = slackMethodPaths[input.Source.Method]; exists == false {
		return "", nil, false, fmt.Errorf("Method '%s' does not exist", input.Source.Method)
	}

	switch input.Source.Method {
	case MethodFilesUpload:
		data, err := ValidateAndBuildPostBodyFilesUpload(input)
		return method, data, false, err
	case MethodChatPostmessage:
		data, empty, err := ValidateAndBuildPostBodyPostMessage(input)
		return method, data, empty, err
	default:
		return "", nil, false, fmt.Errorf("No implementation for building POST body for method '%s'", input.Source.Method)
	}
}

// ValidateAndBuildPostBodyPostMessage ... Processes 'chat.postMessage' request types
func ValidateAndBuildPostBodyPostMessage(input ConcourseInput) (data url.Values, empty bool, err error) {
	data = url.Values{}

	attachmentsStringRaw, err := ValidatePostMessageAttachments(input.Params.Attachments, input.Params.AttachmentsFile)
	if err != nil {
		HandleNonFatalError(err, "Error validating attachments contents")
		attachmentsStringRaw = "[{\"title\": \"[INTERNAL ERROR] Failed to parse string to post to slack\", \"color\": \"danger\"}]"
	}
	attachmentString := ExpandEnv(attachmentsStringRaw)

	switch "" {
	case input.Source.Token:
		return nil, false, fmt.Errorf("Token is a required param")
	case input.Params.Channel:
		return nil, false, fmt.Errorf("Channel is a required param")
	}

	attachments := []Attachment{}
	err = json.Unmarshal([]byte(attachmentString), &attachments)
	if err == nil {
		if len(attachments) != 0 {
			hasText := false
			for _, a := range attachments {
				if (len(a.Text) != 0) || (len(a.Title) != 0) {
					hasText = true
				}
			}
			if !hasText {
				return data, true, err
			}
		}
	}

	data.Set("attachments", attachmentString)
	data.Set("token", input.Source.Token)
	data.Set("channel", input.Params.Channel)
	data.Set("link_names", strconv.Itoa(input.Params.LinkNames))

	if input.Params.IconURL != "" {
		data.Set("icon_url", input.Params.IconURL)
	}
	if input.Params.Username != "" {
		data.Set("username", input.Params.Username)
	}

	return data, false, nil
}

// ValidatePostMessageAttachments ... Validates 'chat.postMessage' request types
func ValidatePostMessageAttachments(attachments string, attachmentsFile string) (string, error) {
	if attachments != "" && attachmentsFile != "" {
		err := fmt.Errorf("cannot supply both attachmentsFile and attachments for Chat.PostMessage")
		return "", err
	}
	if attachments == "" && attachmentsFile == "" {
		err := fmt.Errorf("must supply one of attachmentsFile or attachments for Chat.PostMessage")
		return "", err
	}

	if attachments != "" {
		return attachments, nil
	}

	attachmentsStringRawBytes, err := ioutil.ReadFile(attachmentsFile)
	if err != nil {
		return "", err
	}
	attachmentsStringRaw := string(attachmentsStringRawBytes)
	if attachmentsStringRaw == "" {
		err := fmt.Errorf("attachments file is empty")
		return "", err
	}

	return attachmentsStringRaw, nil
}

// ValidateAndBuildPostBodyFilesUpload ... Processes 'files.upload' request types
func ValidateAndBuildPostBodyFilesUpload(input ConcourseInput) (url.Values, error) {
	data := url.Values{}

	if input.Params.File != "" && input.Params.Content != "" {
		return nil, errors.New("cannot supply both file and content for Files.Upload")
	}
	if input.Params.File == "" && input.Params.Content == "" {
		return nil, errors.New("must supply one of file or content for Files.Upload")
	}

	if input.Params.File != "" {
		fileContents, err := ioutil.ReadFile(input.Params.File)
		if err != nil {
			return data, err
		}
		data.Set("content", string(fileContents))
	} else {
		data.Set("content", input.Params.Content)
	}

	data.Set("token", input.Source.Token)
	data.Set("channels", input.Params.Channels)
	data.Set("title", input.Params.Title)

	return data, nil
}

// PostToSlack ... Sends preconstructed message to post to Slack via Slack's HTTP API
func PostToSlack(path string, data url.Values, fallbackChannel string, tryBackupOnFailure bool) (SlackResponse, error) {
	responseObject := SlackResponse{}

	u, _ := url.ParseRequestURI(SlackAPIURL)
	u.Path = path
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return responseObject, err
	}

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseObject, err
	}
	responseString := string(responseBytes)
	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, responseString)
		return responseObject, fmt.Errorf("Expected 200 response code but got %v", resp.StatusCode)
	}

	parseErr := json.Unmarshal(responseBytes, &responseObject)
	if parseErr != nil {
		return responseObject, parseErr
	}

	if !responseObject.Ok {
		return handleSlackPostResponseNotOkError(responseObject, responseString, fallbackChannel, path, data, tryBackupOnFailure)
	}

	return responseObject, nil
}

func handleSlackPostResponseNotOkError(response SlackResponse, responseString string, fallbackChannel string, path string,
	data url.Values, tryBackupOnFailure bool) (SlackResponse, error) {
	// Is our error related to an invalid channel? Do we have a backup channel defined? If so use it and try again...
	if (response.Error == "channel_not_found" || response.Error == "invalid_channel" || response.Error == "is_archived") &&
		strings.TrimSpace(fallbackChannel) != "" && tryBackupOnFailure {
		if data.Get("channel") != "" {
			data.Set("channel", fallbackChannel)
			return PostToSlack(path, data, fallbackChannel, false)
		} else if data.Get("channels") != "" {
			data.Set("channels", fallbackChannel)
			return PostToSlack(path, data, fallbackChannel, false)
		}
	}

	fmt.Fprintf(os.Stderr, "%s\n", responseString)
	return response, errors.New("Slack API returned 'ok': false ")
}

// ExpandEnv ... Wrapper for os.ExpandEnv that prevents single quoted strings of the type '$something' from being interpreted as env variables
func ExpandEnv(s string) string {
	valueDictionary := make(map[string]string)

	r, _ := regexp.Compile("'\\$[a-zA-Z0-9]+'")

	uniqueValues := RemoveDuplicatesUnordered(r.FindAllString(s, -1))

	for index, match := range uniqueValues {
		if _, exists := valueDictionary[match]; !exists {
			valueDictionary[match] = "!?!?" + strconv.Itoa(index) + "!?!?"
		}
		s = strings.Replace(s, match, valueDictionary[match], -1)
	}

	s = os.ExpandEnv(s)

	for key := range valueDictionary {
		s = strings.Replace(s, valueDictionary[key], key, -1)
	}
	return s
}

// RemoveDuplicatesUnordered ... Essentially converts a list to a set
func RemoveDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	for v := range elements {
		encountered[elements[v]] = true
	}

	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}
