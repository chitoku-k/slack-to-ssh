slack-to-ssh
============

[![][workflow-badge]][workflow-link]

Runs an SSH command from Slack interactive message buttons.

## Requirements

- Go
- Slack App

## Installation

```sh
$ docker buildx build .
```

```sh
# Port number (required)
export PORT=8080

# TLS certificate and private key (optional; if not specified, application is served over HTTP)
export TLS_CERT=/path/to/tls/cert
export TLS_KEY=/path/to/tls/key

# SSH Hostname (required)
export SSH_HOSTNAME=

# SSH Port (optional; defaults to 22)
export SSH_PORT=

# SSH Username (required)
export SSH_USERNAME=

# SSH Private Key (required)
export SSH_PRIVATE_KEY_FILE=/path/to/private/key

# Slack App Secret (required)
export SLACK_APP_SECRET=

# nth Action Name (required)
export SLACK_ACTION_0_NAME=

# nth Attachment Text (optional)
export SLACK_ACTION_0_ATTACHMENT_TEXT=

# nth SSH Command to execute (required)
export SLACK_ACTION_0_COMMAND=
```

## Usage

1. Send a message to Slack with actions:

```json
{
    "text": "Click a button!",
    "attachments": [
        {
            "title": "Questionnaire",
            "text": "What's for dinner?",
            "actions": [
                {
                    "name": "exec",
                    "type": "button",
                    "text": "Beef",
                    "value": "beef"
                },
                {
                    "name": "exec",
                    "type": "button",
                    "style": "danger",
                    "text": "Turkey",
                    "value": "turkey",
                    "confirm": {
                        "title": "Eating turkey?",
                        "text": "Are you sure you want to eat turkey?",
                        "ok_text": "OK",
                        "dismiss_text": "Cancel"
                    }
                }
            ]
        }
    ]
}
```

2. Serve the action with the following config:

```sh
export SSH_HOSTNAME=example.com
export SSH_USERNAME=user
export SSH_PRIVATE_KEY_FILE=/path/to/private/key
export SLACK_APP_SECRET=00000000000000000000000000000000
export SLACK_ACTION_0_NAME=beef
export SLACK_ACTION_0_ATTACHMENT_TEXT='Beef is chosen for dinner'
export SLACK_ACTION_0_COMMAND='echo Bonjour, beef | cowsay'
export SLACK_ACTION_1_NAME=turkey
export SLACK_ACTION_1_ATTACHMENT_TEXT='Turkey chosen for dinner'
export SLACK_ACTION_1_COMMAND='echo Hello, turkey | cowsay -f turkey'
```

3. The specified command is executed on the remote server.

[workflow-link]:    https://github.com/chitoku-k/slack-to-ssh/actions?query=branch:master
[workflow-badge]:   https://img.shields.io/github/workflow/status/chitoku-k/slack-to-ssh/CI%20Workflow/master.svg?style=flat-square
