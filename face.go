package greenyfy

import (
    "bytes"
    "net/http"
    
    "image"
    "image/jpeg"
    
    "appengine"
    "appengine/urlfetch"
    
    "encoding/json"
    "errors"
)

func findFaces(c appengine.Context, img *image.Image) ([]Face, error) {
    client := urlfetch.Client(c)
    endpoint := config.FaceApiUrl + "?returnFaceLandmarks=true&returnFaceAttributes=headPose"
    bodyType := "application/octet-stream"
    
    buffer := new(bytes.Buffer)
    if err := jpeg.Encode(buffer, *img, nil); err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", endpoint, buffer)
    
    if err != nil {
        return nil, err
    }

    req.Header.Add("Content-Type", bodyType)
    req.Header.Add("Ocp-Apim-Subscription-Key", config.FaceApiKey)
    
    resp, err := client.Do(req)

    if err != nil {
        return nil, err
    }
        
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        c.Errorf("Failed to access FaceAPI, return code: %v", resp.Status)
        return nil, errors.New("Could not access API")
    }

    var obj []Face
    dec := json.NewDecoder(resp.Body)
    err = dec.Decode(&obj)
    
    if err != nil {
        return nil, err
    }
        
    return obj, nil
}

type Face struct {
    Id   string `json:"faceId"`
    Rectangle FaceRectangle `json:"faceRectangle"`
    Landmarks FaceLandmarks `json:"faceLandmarks"`
    Attributes FaceAttributes
}

type FaceLandmarks struct {
    PupilLeft Point
    PupilRight Point
    NoseTip Point
    MouthLeft Point
    MouthRight Point
    EyebrowLeftOuter Point
    EyebrowLeftInner Point
    EyeLeftOuter Point
    EyeLeftTop Point
    EyeLeftBottom Point
    EyeLeftInner Point
    EyebrowRightInner Point
    EyebrowRightOuter Point
    EyeRightInner Point
    EyeRightTop Point
    EyeRightBottom Point
    EyeRightOuter Point
    NoseRootLeft Point
    NoseRootRight Point
    NoseLeftAlarTop Point
    NoseRightAlarTop Point
    NoseLeftAlarOutTip Point
    NoseRightAlarOutTip Point
    UpperLipTop Point
    UpperLipBottom Point
    UnderLipTop Point
    UnderLipBottom Point
}

type Point struct {
    X float32
    Y float32
}

type FaceRectangle struct {
    Top float32
    Left float32
    Width float32
    Height float32
}

type FaceAttributes struct {
    Pose HeadPose `json:"headPose"`
}

type HeadPose struct {
    Pitch float32
    Roll float32
    Yaw float32
}
