# Relay slack app

# Note
    - always replace the `request url` with the new ngrok url when a new ngrok started

# Before production
    - replace the `request url` in the events page with the new domain.
    - replace the `slack_signing_secret` or add it in the ENV
    - add ip-address of the server in the `Restrict API Token Usage` in slack's OAuth page

# References
    - https://github.com/slack-go/slack
