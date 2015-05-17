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
    endpoint := "https://api.projectoxford.ai/face/v0/detections?analyzesFaceLandmarks=true&analyzesHeadPose=true"
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
/*
    "attributes": {
        "headPose": {
            "pitch": 0.0,
            "roll": -4.7,
            "yaw": -1.5
        }
    }
*/