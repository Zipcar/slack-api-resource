package common

// ConcourseSource ... Defines the format for the 'source' stanza received as input for the resource
type ConcourseSource struct {
	Token  string `json:"token"`
	Method string `json:"method"`
}

// ConcourseParams ... Defines the format for the 'params' stanza received as input for the resource
type ConcourseParams struct {
	// common Params
	FallbackChannel string `json:"fallback_channel"`

	//fileUpload Params
	Content  string `json:"content"`
	File     string `json:"file"`
	Title    string `json:"title"`
	Channels string `json:"channels"`

	//postMessage Params
	Channel         string `json:"channel"`
	AttachmentsFile string `json:"attachments_file"`
	Attachments     string `json:"attachments"`
	IconURL         string `json:"icon_url"`
	Username        string `json:"username"`
	LinkNames       int    `json:"link_names"`
}

// ConcourseInput ... Defines the overall expected input format for the resource
type ConcourseInput struct {
	Source  ConcourseSource  `json:"source"`
	Params  ConcourseParams  `json:"params"`
	Version ConcourseVersion `json:"version"`
}

// Attachment ... Defines the output format for a single attachment on a Slack message
type Attachment struct {
	Text  string `json:"text"`
	Title string `json:"title"`
}

// ConcourseVersion ... Defines the format for the 'version' stanza received as input for the resource
type ConcourseVersion struct {
	Ref string `json:"ref"`
}

// SlackResponse ... Defines the output format from the Slack API
type SlackResponse struct {
	Ok   bool              `json:"ok"`
	File SlackFileResponse `json:"file"`
	Error string           `json:error`
}

// SlackFileResponse ... Defines the output format for the Slack file portion of the Slack API response
type SlackFileResponse struct {
	ID string `json:"id"`
}
