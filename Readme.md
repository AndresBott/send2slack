# send2slack

Simple command line applications to send messages to slack.
The motivation is to create an almost drop-in replacement for `sendmail` to send slack messages.

send2slack has some specific attributes that can be used to enhance the messages you send.

# How to use it

before you use send2slack, create an App in https://api.slack.com/apps and install it in the desired workspace

## Config
send2slack looks for a config.yaml file in the following locations in this order

   * /etc/send2slack/config.yaml
   * $HOME/.send2slack/config.yaml
   * ./config.yaml
   
the configuration for now is quite simple and only contains:

    ---
    token: "<slack token goes here>"
    default_channel: "general"
    sendmail_channel: "mails"
    
### slack app
this is the minimum needed to create the oauth token

1 go to https://api.slack.com/apps

2 Create a new app with the following Scopes:
  - Bot Token Scopes:
      - chat:write
      - chat:write.customize
      - chat:write.public
      
3 use the "Bot User OAuth Access Token" in the configuration 

# Sample usage

send a simple image

    send2slack "this is a message"
    
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

more complex messages can be sent by inputing a file or a "[Here Documents](https://tldp.org/LDP/abs/html/here-docs.html)" construct 

loading the sample file:

    send2slack < sampleMsg.md
    
here documents:

    send2slack <<EOF
    > This is a sample message to be sent as payload to slack using send2slack.
    > When sending payloads you can use: 
    > • *bold text*
    > • _italic text_
    EOF
    
## sendmail replacement

if the binary is renamed to sendmail ( or symlink to a renamed ) the command line parameters will be ignored and it 
will accept emails on stdin that will be forwarded to channel configured with "sendmail_channel"