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

const (
	SlackApiUrl           = "https://slack.com"
	MethodFilesUpload     = "files.upload"
	MethodChatPostmessage = "chat.postMessage"
)

var slackMethodPaths = map[string]string{
	MethodFilesUpload:     "api/files.upload",
	MethodChatPostmessage: "api/chat.postMessage",
}

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

func HandleFatalError(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err.Error()))
	os.Exit(1)
}

func HandleNonFatalError(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err.Error()))
}

func ValidateAndBuildPostBody(input ConcourseInput) (string, url.Values, error) {
	var method = ""
	var exists = false
	if method, exists = slackMethodPaths[input.Source.Method]; exists == false {
		return "", nil, errors.New(fmt.Sprintf("Method '%s' does not exist", input.Source.Method))
	}

	switch input.Source.Method {
	case MethodFilesUpload:
		data, err := ValidateAndBuildPostBodyFilesUpload(input)
		return method, data, err
	case MethodChatPostmessage:
		data, err := ValidateAndBuildPostBodyPostMessage(input)
		return method, data, err
	default:
		return "", nil, errors.New(fmt.Sprintf("No implementation for building POST body for method '%s'", input.Source.Method))
	}
}

func ValidateAndBuildPostBodyPostMessage(input ConcourseInput) (url.Values, error) {
	data := url.Values{}

	attachmentsStringRaw, err := ValidatePostMessageAttachments(input.Params.Attachments, input.Params.AttachmentsFile)
	if err != nil {
		HandleNonFatalError(err, "Error validating attachments contents")
		attachmentsStringRaw = "[{\"title\": \"[INTERNAL ERROR] Failed to parse string to post to slack\", \"color\": \"danger\"}]"
	}
	attachmentString := ExpandEnv(attachmentsStringRaw)

	switch "" {
	case input.Source.Token:
		return nil, errors.New(fmt.Sprintf("Token is a required param"))
	case input.Params.Channel:
		return nil, errors.New(fmt.Sprintf("Channel is a required param"))
	}

	data.Set("attachments", attachmentString)
	data.Set("token", input.Source.Token)
	data.Set("channel", input.Params.Channel)
	data.Set("link_names", strconv.Itoa(input.Params.LinkNames))

	if input.Params.IconUrl != "" {
		data.Set("icon_url", input.Params.IconUrl)
	}
	if input.Params.Username != "" {
		data.Set("username", input.Params.Username)
	}

	return data, nil
}

func ValidatePostMessageAttachments(attachments string, attachmentsFile string) (string, error) {
	if attachments != "" && attachmentsFile != "" {
		err := errors.New(fmt.Sprintf("cannot supply both attachmentsFile and attachments for Chat.PostMessage"))
		return "", err
	}
	if attachments == "" && attachmentsFile == "" {
		err := errors.New(fmt.Sprintf("must supply one of attachmentsFile or attachments for Chat.PostMessage"))
		return "", err
	}

	if attachments != "" {
		return attachments, nil
	} else {
		attachmentsStringRawBytes, err := ioutil.ReadFile(attachmentsFile)
		if err != nil {
			return "", err
		}
		attachmentsStringRaw := string(attachmentsStringRawBytes)
		if attachmentsStringRaw == "" {
			err := errors.New(fmt.Sprintf("attachments file is empty"))
			return "", err
		}
		return attachmentsStringRaw, nil
	}
}

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

func PostToSlack(path string, data url.Values) (SlackResponse, error) {
	responseObject := SlackResponse{}

	u, _ := url.ParseRequestURI(SlackApiUrl)
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
		return responseObject, errors.New(fmt.Sprintf("Expected 200 response code but got %v", resp.StatusCode))
	}

	parseErr := json.Unmarshal(responseBytes, &responseObject)
	if parseErr != nil {
		return responseObject, parseErr
	}
	if !responseObject.Ok {
		fmt.Fprintf(os.Stderr, "%s\n", responseString)
		return responseObject, errors.New("Slack API returned 'ok': false ")
	}

	return responseObject, nil
}

// a wrapper for os.ExpandEnv that prevents single quoted strings
// of the type '$something' from being interpreted as env variables
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
