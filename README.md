# Slack API Resource

A [Concourse](https://concourse-ci.org/) resource for using the [slack API](https://api.slack.com/) implemented using either the legacy token or OAuth2 token returned when creating a slack app.

Note that this differs from the [slack-notification-resource](https://github.com/cloudfoundry-community/slack-notification-resource)
as it uses the slack API rather than the slack webhook, therefore allowing for use of all slack API functionality and not just message posting.

## Authentication

There is currently no ability to execute the OAuth2 workflow to create a token.  Instead, you are expected to pass a working token.  Three options:
1) Use a 'legacy' token - https://api.slack.com/custom-integrations/legacy-tokens
1) Using a slack app, set permission scopes and then use the OAuth Access Token
1) Using a slack app, previous to invoking the resource, follow the OAuth2 workflow to get a token - https://api.slack.com/docs/oauth

## Source Configuration

1) method - See list of methods
2) token - See Authentication section

Example
```
resource_types:
- name: slack-file-upload
  source:
    token: REDACTED
    method: files.upload
  type: slack-api
```

### List of Available Methods

This resource has not implemented all available slack API methods.  Feel free to contribute!

1) files.upload - https://api.slack.com/methods/files.upload
1) chat.postMessage - https://api.slack.com/methods/chat.postMessage

## Parameter Configuration

### Files.Upload 

Currently only supports text data.

1) channels - comma separated list of channels
1) content - the text that will be sent to the given channel(s) as a snippet and also stored as a file
1) title - the title of the snippet and file
1) file - the file, passed from a task or resource in the job, whose contents will be uploaded

Note that you can ONLY use `file` or `content`, not both

### Chat.PostMessage

Currently only supports posting `attachments` property in the slack message, and not other top level properties.

1) channel - specify what channel to post to
1) attachments_file - the array of attachments to post, passed in as a JSON file from a task or resource in the job
1) attachments - the array of attachments to post
1) icon_url (optional) - url for the icon to use in post
1) username (optional) - user to post as

Note that you can ONLY use `attachments_file` or `attachments`, not both

Example attachments_file would be a path to a file with the following contents.  
Or you could pass this in as the value of attachments directly:
```
[
  { 
  	"title": "test attachment 1", 
  	"text": "test attachment 1 text" 
  }
]
```

## Contributing

### Building the Image and Running Tests

Note that the tests are run as part of the Docker build, and the build will fail if any tests fail.

    docker build -t slack-api-resource .

You can also test with the script end-to-end by starting that container in interactive mode via

    docker run -it slack-api-resource bash

and then calling the script of your choice. As an example, to execute posting a message to some slack channel:

    cd /opt/resource
    export SLACK_TOKEN=REDACTED
    export TEST_SLACK_CHANNEL=REDACTED
    echo '[{ "title": "test attachment 1", "text": "test attachment 1 text" }]' > msg_attachments.json
    echo "{\"source\": { \"token\" :\"${SLACK_TOKEN}\", \"method\": \"chat.postMessage\" }, \"params\": { \"attachments_file\": \"msg_attachments.json\", \"channel\" : \"${TEST_SLACK_CHANNEL}\", \"icon_url\": \"http://cl.ly/image/3e1h0H3H2s0P/concourse-logo.png\", \"username\": \"concourse\"}}" | ./out .

### Publishing to public docker registry

Currently you'll need to build and push to your local registry until we build support for this in the public docker registry.