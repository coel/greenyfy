package greenyfy

var config = struct {
		FaceApiUrl string
		FaceApiKey string
		BeardUrl string
	} {
		FaceApiUrl: "https://api.projectoxford.ai/face/v0/detections",
		FaceApiKey: "",
		BeardUrl: "http://greenyfy.appspot.com/images/beard.png",
	}
