# send2slack

Simple command line applications to send messages to slack.
The motivation for this project is to route as many notifications as possible from email to slack.

send2slack is a polymorphic binary that will behave differently depending on the passed arguments.

# How to use it

before you use send2slack, create an App in https://api.slack.com/apps and install it in the desired workspace
## slack app
this is the minimum needed to create the oauth token

1 go to https://api.slack.com/apps

2 Create a new app with the following Scopes:
  - Bot Token Scopes:
      - chat:write
      - chat:write.customize
      - chat:write.public
      
3 use the "Bot User OAuth Access Token" in the configuration 

## Config
send2slack uses two configuration files, sever.yaml and client.yaml, the search paths are: 

* ./
* $HOME/.send2slack/
* /etc/send2slack/

see the sample configurations in resources/config for configuration details
    
# Client mode usage

This mode uses client.yaml as configuration file.

## Direct mode

In direct mode send2slack reads the slack token directly from a configuration file or the env variable $SLACK_TOKEN
and sends the message to slack.

the drawback is that in order to use this all users need to have read access to the config file that contains the 
token; this is not good if used as system wide installation. 

## Proxy mode

In Proxy mode, send2slack send a http message to it's server counter part, and the server then delivers the message 
to slack.

In this model, since only the server needs access to the token, this access can be limited to only one user.  
    
## flags


`-c --color [#xxxxxx | red | green | blue | orange | lime ] ` add a colored block to the message

        send2slack -C red "this is a message"

`-d, --channel <channel> `  channel to send the message, de default is specified in the configuration file

## formatting messages

when sending messages, the formatting is passed to the api, see `sampleMsg.md` for some samples or check 
https://api.slack.com/reference/surfaces/formatting for more details.

samples:

    send2slack 'send an @here: <!here>' // note the single quote
    send2slack 'emojis : :slightly_smiling_face:'
    
## complex messages

more complex messages can be sent by piping a file or a "[Here Documents](https://tldp.org/LDP/abs/html/here-docs.html)" construct 

loading the sample file:

    send2slack < sampleMsg.md
    
here documents:

    send2slack <<EOF
    > This is a sample message to be sent as payload to slack using send2slack.
    > When sending payloads you can use: 
    > • *bold text*
    > • _italic text_
    EOF
    
# Daemon mode

This mode uses server.yaml as configuration file.

## Server

In server mode send2slack will start an unauthenticated http server that accepts post requests from the client.

In order to start the in server mode the configuration field `listen_url` has to be different from `false` the binary
invoked with the flag `-s`

    send2slack -s -f /my/config/file.yaml 

## mbox watcher

In server mode send2slack will watch file modifications on the directory specified with in `mbox_watch` and consume 
the emails written to these files delivering them as slack messages

In order to start the in server mode the configuration field `mbox_watch` has to be different from `false` the binary
invoked with the flag `-w`

    send2slack -w -f /my/config/file.yaml 
