Greenyfy
========

Provide an image url and have beards added to all the people.

Inspired by mustachify.me and one man's epic beard.

Requirements
------------

Built to run on Google App Engine.  The Google App Engine SDK for Go is required to run this: [https://cloud.google.com/appengine/downloads](https://cloud.google.com/appengine/downloads)

Facial recognition is done via Microsoft Project Oxford Face API, sign up at: [http://www.projectoxford.ai/face](http://www.projectoxford.ai/face)

To run, you need to set FaceAPIKey to config.go with an API key from Microsoft after signing up above.

Running
-------

With the Google App Engine SDK for Go installed, run in the greenyfy folder:

> goapp serve

The app will then be available at:

> localhost:8080
