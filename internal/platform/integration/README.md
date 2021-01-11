# How access works
The access part is straight forward. First we download the credentials.json from the google console credentials for OAuth.
Using the OAuth file we generate the auth-url and pass it to the UI. 
While creating the OAuth credentials we will also give the callback/redirect URL. (For localhost, we gave localhost:4200/auth/handler as the redirect URL).
The UI will pass the code to the server end and closes the window.
The Server will store the token and call the watch with the topic name. So, that the app will start receving the message on the particular topic.

# How receive works
- We create a topic in the google console called "receive-gmail-message" and also we created the subscription called "sub-receive-gmail-message" 
- On subscription we configured the push method which sends the message to the path `/receive/gmail/message` 
- Since we can't give the localhost here. We are running the ngRok to receive the callback. 
- ngrok http -hostname=vjrelay.ngrok.io 3000

# Todo
- Add authentication for the push URL
- Process the received message

# Production Todo
- Create a new prod topic and give that topic to the config
- Add the prod callback url in the subscription
- Add the prod OAuth credentials and the callback URL
