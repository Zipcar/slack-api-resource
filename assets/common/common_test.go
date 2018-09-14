package common_test

import (
	"fmt"
	"testing"

	"../common"
)

func TestValidateAndBuildBody_FilesUpload_Content(t *testing.T) {
	input := common.ConcourseInput{}
	input.Source.Method = common.MethodFilesUpload
	input.Params.Content = "my content"
	input.Params.Title = "my title"
	input.Params.Channels = "mychannel1,mychannel2"
	input.Source.Token = "validToken"

	expectedMethod := "api/" + input.Source.Method
	expectedData := map[string]string{
		"title":    input.Params.Title,
		"token":    input.Source.Token,
		"content":  input.Params.Content,
		"channels": input.Params.Channels,
	}

	method, data, err, _ := common.ValidateAndBuildPostBody(input)

	if err != nil {
		t.Errorf("Unexpected error, got: %s, want: %v.", err.Error(), nil)
	}

	if method != expectedMethod {
		t.Errorf("Unexpected method, got: %s, want: %s.", method, expectedMethod)
	}

	for key, value := range data {
		var expectedValue = ""
		var exists = false
		if expectedValue, exists = expectedData[key]; exists == false {
			t.Errorf("Unexpected data key: %s.", key)
		} else if value[0] != expectedValue {
			t.Errorf("Unexpected data value, got: %s, want: %s.", value[0], expectedValue)
		}
	}
	for expectedKey := range expectedData {
		if _, exists := data[expectedKey]; exists == false {
			t.Errorf("Missing expected data key: %s.", expectedKey)
		}
	}
}

func TestValidatePostMessageAttachments_AttachmentsSuccess(t *testing.T) {
	expectedAttachmentString := "[{'text':'fun'}]"
	actualAttachmentsString, actualErr := common.ValidatePostMessageAttachments(expectedAttachmentString, "")

	if actualErr != nil {
		t.Errorf("Unexpected error, got: %s, want: %v.", actualErr.Error(), nil)
	}

	if expectedAttachmentString != actualAttachmentsString {
		t.Errorf("Unexpected error, got: %s, want: %v.", actualAttachmentsString, expectedAttachmentString)
	}
}

func TestValidatePostMessageAttachments_MissingAttachmentsAndFile(t *testing.T) {
	_, actualErr := common.ValidatePostMessageAttachments("", "")

	if actualErr == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestValidatePostMessageAttachments_BothAttachmentsAndFile(t *testing.T) {
	_, actualErr := common.ValidatePostMessageAttachments("test1", "test2")

	if actualErr == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestMessage(t *testing.T) {
	input := common.ConcourseInput{}
	input.Source.Method = "chat.postMessage"
	input.Params.AttachmentsFile = "example.json"
	input.Source.Token = "validToken"
	input.Params.Channel = "mychannel1"
	_, data, err, _ := common.ValidateAndBuildPostBody(input)

	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if len(data.Get("attachments")) == 0 {
		t.Errorf("Expected attahments to exist")
	}
}
func TestEmptyMessage(t *testing.T) {
	input := common.ConcourseInput{}
	input.Source.Method = "chat.postMessage"
	input.Params.AttachmentsFile = "empty.json"
	input.Source.Token = "validToken"
	input.Params.Channel = "mychannel1"
	_, _, err, empty := common.ValidateAndBuildPostBody(input)

	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if !empty {
		t.Error("ValidateAndBuildPostBody didnt return an empty message of true when sent an emtpy message")
	}
}
func TestValidateAndBuildBody_InvalidMethod(t *testing.T) {
	input := common.ConcourseInput{}
	input.Source.Method = "invalid.method"

	expectedErrorMessage := fmt.Sprintf("Method '%s' does not exist", input.Source.Method)

	_, _, err, _ := common.ValidateAndBuildPostBody(input)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if err != nil && err.Error() != expectedErrorMessage {
		t.Errorf("Unexpected error, got: %s, want: %s.", err.Error(), expectedErrorMessage)
	}
}
