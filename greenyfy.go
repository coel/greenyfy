package greenyfy

import (
    "bytes"
    "fmt"
    "log"
    "net/http"
	"math"
    
    "image"
    "image/jpeg"
    _ "image/png"
    "image/draw"
	//"golang.org/x/image/draw"
	//"golang.org/x/image/math/f64"

    "strconv"
    "github.com/nfnt/resize"
    
    "appengine"
    "appengine/urlfetch"
    
    "encoding/json"
	
	"code.google.com/p/graphics-go/graphics"
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
    if bnds.Dx() > 1024 {
        log.Println("Resizing image", bnds.Dx())
        img = resize.Resize(1024, 0, img, resize.Lanczos3)
    }
    
	faces := findFaces(c, &img)
	// todo: should I pass back by reference?
	
	log.Println("Obj: ", faces)
		
    bnds = img.Bounds()

    m := image.NewRGBA(image.Rect(0, 0, bnds.Dx(), bnds.Dy()))
    
    draw.Draw(m, bnds, img, image.Point{0,0}, draw.Src)
    
    brd := beard(c)
	
	for _, face := range faces {
		
		vert := int((face.Landmarks.NoseTip.Y + face.Landmarks.NoseTip.Y + face.Landmarks.UpperLipTop.Y) / 3)
	    brd_resized := resize.Resize(uint(face.Rectangle.Width), 0, brd, resize.Lanczos3)
		brd_bnds := brd_resized.Bounds()
		
	
		//draw.Draw(m, rt, brd_resized, sr.Min, draw.Over)

//draw.Draw(m, m.Bounds(), img2, image.Point{-200,-200}, draw.Src)
		rb := image.NewRGBA(image.Rect(0, 0, brd_bnds.Dx(), brd_bnds.Dy()))
    
		rad := float64(face.Attributes.Pose.Roll)*math.Pi/180
    	graphics.Rotate(rb, brd_resized, &graphics.RotateOptions{rad})

		t := math.Tan(rad) * float64(brd_bnds.Dy()) / 2

		log.Println("Offset: ", t)

	    sr := image.Rect(0,0,brd_bnds.Dx()*4,brd_bnds.Dy()*4)
	    dp := image.Point{int(face.Rectangle.Left) - int(t), vert}
	    rt := image.Rectangle{dp, dp.Add(sr.Size())}

		draw.Draw(m, rt, rb, sr.Min, draw.Over)

		//matrix := f64.Aff3{1, -0.2, 0, 1, 0, 0}
	    //draw.NearestNeighbor.Transform(m, &matrix, brd_resized, rt, draw.Over, nil)

	} 
	
    
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

func findFaces(c appengine.Context, img *image.Image) []Face {
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
		
	return obj
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