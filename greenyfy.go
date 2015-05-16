package greenyfy

import (
    "bytes"
    "fmt"
    "log"
    "net/http"
    
    "image"
    "image/jpeg"
    _ "image/png"
    "image/draw"
    "strconv"
    "github.com/nfnt/resize"
    
    "appengine"
    "appengine/urlfetch"
    
    "encoding/json"
)

func init() {
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    img_url := r.FormValue("me")

    if len(img_url) == 0 {
        fmt.Fprint(w, "Because API")
        return
    }
    
    c := appengine.NewContext(r)
    client := urlfetch.Client(c)
    resp, err := client.Get(img_url)

    if err != nil {
        log.Println("Failed to get url: ", img_url)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    defer resp.Body.Close()
    
    // fmt.Fprintf(w, "HTTP GET returned status %v", resp.Status)
    
    img, _, err := image.Decode(resp.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    bnds := img.Bounds()
    if bnds.Max.X > 1024 {
        log.Println("Resizing image", bnds.Max.X)
        img = resize.Resize(1024, 0, img, resize.Lanczos3)
    }
    
	findFaces(c, &img)
	
    bnds = img.Bounds()
    
    brd := beard(c)
    //log.Println("beard: ", brd)
    
    sr := image.Rect(110,110,500,550)
    dp := image.Point{10, 10}
    rt := image.Rectangle{dp, dp.Add(sr.Size())}
    m := image.NewRGBA(image.Rect(0, 0, bnds.Max.X, bnds.Max.Y))
    
    draw.Draw(m, bnds, img, image.Point{0,0}, draw.Src)
    draw.Draw(m, rt, brd, sr.Min, draw.Over)
    
    img_out := image.Image(m)
    
    writeImage(w, &img_out)
}

// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writeImage(w http.ResponseWriter, img *image.Image) {

    buffer := new(bytes.Buffer)
    if err := jpeg.Encode(buffer, *img, nil); err != nil {
        log.Println("unable to encode image.")
    }

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
    if _, err := w.Write(buffer.Bytes()); err != nil {
        log.Println("unable to write image.")
    }
}

func beard(c appengine.Context) image.Image {
    client := urlfetch.Client(c)
    resp, err := client.Get("http://localhost:8080/images/beard.png")

    if err != nil {
        log.Println("Failed to get beard url")
    }
    
    defer resp.Body.Close()
    
    img, _, _ := image.Decode(resp.Body)
    
    return img
}

func findFaces(c appengine.Context, img *image.Image) {
    client := urlfetch.Client(c)
    endpoint := "https://api.projectoxford.ai/face/v0/detections?analyzesFaceLandmarks=true"
    bodyType := "application/octet-stream"
    
    buffer := new(bytes.Buffer)
    if err := jpeg.Encode(buffer, *img, nil); err != nil {
        log.Println("unable to encode image.")
    }
    
    req, err := http.NewRequest("POST", endpoint, buffer)
    req.Header.Add("Content-Type", bodyType)
    req.Header.Add("Ocp-Apim-Subscription-Key", "b8fbe571719248759e2d2badff0d1d6f")
    resp, err := client.Do(req)

    if err != nil {
        log.Println("Failed to get beard url")
    }
    	
    defer resp.Body.Close()

	obj := make([]Face, 1)
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&obj)
	
	log.Println("Obj: ", obj)
	log.Println("Pupil: ", obj[0].Landmarks.PupilLeft)
}

type Face struct {
    Id   string `json:"faceId"`
	Rectangle FaceRectangle `json:"faceRectangle"`
    Landmarks FaceLandmarks `json:"faceLandmarks"`
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

/*
[{
    "faceId": "c2d9a455-c894-4256-a2a7-01924307b94b",
    "faceRectangle": {
        "top": 0,
        "left": 211,
        "width": 185,
        "height": 181
    },
    "faceLandmarks": {
        "pupilLeft": {
            "x": 262.1,
            "y": 40.8
        },
        "pupilRight": {
            "x": 344.2,
            "y": 44.6
        },
        "noseTip": {
            "x": 306.5,
            "y": 94.1
        },
        "mouthLeft": {
            "x": 272.0,
            "y": 133.5
        },
        "mouthRight": {
            "x": 330.5,
            "y": 133.8
        },
        "eyebrowLeftOuter": {
            "x": 236.0,
            "y": 24.1
        },
        "eyebrowLeftInner": {
            "x": 290.9,
            "y": 33.3
        },
        "eyeLeftOuter": {
            "x": 247.2,
            "y": 43.5
        },
        "eyeLeftTop": {
            "x": 260.6,
            "y": 35.4
        },
        "eyeLeftBottom": {
            "x": 260.7,
            "y": 51.1
        },
        "eyeLeftInner": {
            "x": 275.3,
            "y": 44.9
        },
        "eyebrowRightInner": {
            "x": 322.3,
            "y": 33.0
        },
        "eyebrowRightOuter": {
            "x": 372.7,
            "y": 30.9
        },
        "eyeRightInner": {
            "x": 330.3,
            "y": 47.4
        },
        "eyeRightTop": {
            "x": 345.9,
            "y": 39.0
        },
        "eyeRightBottom": {
            "x": 346.0,
            "y": 54.1
        },
        "eyeRightOuter": {
            "x": 360.3,
            "y": 47.9
        },
        "noseRootLeft": {
            "x": 294.7,
            "y": 48.1
        },
        "noseRootRight": {
            "x": 317.1,
            "y": 48.9
        },
        "noseLeftAlarTop": {
            "x": 288.7,
            "y": 80.2
        },
        "noseRightAlarTop": {
            "x": 320.5,
            "y": 83.3
        },
        "noseLeftAlarOutTip": {
            "x": 280.9,
            "y": 96.0
        },
        "noseRightAlarOutTip": {
            "x": 323.9,
            "y": 100.0
        },
        "upperLipTop": {
            "x": 303.0,
            "y": 123.5
        },
        "upperLipBottom": {
            "x": 302.9,
            "y": 130.3
        },
        "underLipTop": {
            "x": 306.0,
            "y": 138.1
        },
        "underLipBottom": {
            "x": 304.1,
            "y": 149.6
        }
    },
    "attributes": {}
}]
*/