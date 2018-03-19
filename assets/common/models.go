package common

type ConcourseSource struct {
	Token    string `json:"token"`
	Method   string `json:"method"`
}

type ConcourseParams struct {
	//fileUpload Params
	Content         string `json:"content"`
	File            string `json:"file"`
	Title           string `json:"title"`
	Channels        string `json:"channels"`

	//postMessage Params
	Channel         string `json:"channel"`
	AttachmentsFile string `json:"attachments_file"`
	IconUrl         string `json:"icon_url"`
	Username        string `json:"username"`
}

type ConcourseInput struct {
	Source  ConcourseSource  `json:"source"`
	Params  ConcourseParams  `json:"params"`
	Version ConcourseVersion `json:"version"`
}

type ConcourseVersion struct {
	Ref  string `json:"ref"`
}

type SlackResponse struct {
	Ok	bool `json:"ok"`
	File SlackFileResponse `json:"file"`
}

type SlackFileResponse struct {
	Id  string `json:"id"`
}