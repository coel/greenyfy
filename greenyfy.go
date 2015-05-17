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

    "strconv"
    "github.com/nfnt/resize"
    
    "appengine"
    "appengine/urlfetch"
	"appengine/memcache"
    
    "encoding/json"
	
	"code.google.com/p/graphics-go/graphics"
	//"io"
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
	
	item, err := memcache.Get(c, img_url)
	
	if err == memcache.ErrCacheMiss {
	    c.Infof("item not in the cache")
		
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
		log.Println("----")
		for _, face := range faces {
			
			brd_resized := resize.Resize(uint(face.Rectangle.Width*2), 0, brd, resize.Lanczos3)
			brd_bnds := brd_resized.Bounds()
			
			vert := (face.Landmarks.MouthLeft.Y + face.Landmarks.MouthRight.Y) /2 - float32(brd_bnds.Dy()) * 0.5
	
			rb := image.NewRGBA(image.Rect(0, 0, brd_bnds.Dx(), brd_bnds.Dy()))
	    
			rad := float64(face.Attributes.Pose.Roll)*math.Pi/180
	    	graphics.Rotate(rb, brd_resized, &graphics.RotateOptions{rad})
	
			mid := face.Rectangle.Left + face.Rectangle.Width / 2 // face.Landmarks.NoseTip.X
			lt := mid - (float32(brd_bnds.Dx()) / 2) // + float32(t)
		    sr := image.Rect(0,0,brd_bnds.Dx()*4,brd_bnds.Dy()*4)
		    dp := image.Point{int(float64(lt)), int(float64(vert))}
		    rt := image.Rectangle{dp, dp.Add(sr.Size())}
	
			draw.Draw(m, rt, rb, sr.Min, draw.Over)
		} 
	log.Println("***")
	    img_out := image.Image(m)
	    
	    buffer := new(bytes.Buffer)
	    if err := jpeg.Encode(buffer, img_out, nil); err != nil {
	        log.Println("unable to encode image.")
	    }

		item = &memcache.Item{
		    Key:   img_url,
		    Value: buffer.Bytes(),
		}
		
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		    c.Infof("item with key %q already exists", item.Key)
		} else if err != nil {
		    c.Errorf("error adding item: %v", err)
		}
	} else if err != nil {
	    c.Errorf("error getting item: %v", err)
	} else {
		log.Println("Cache hit: ", img_url)
	}

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(item.Value)))
    if _, err := w.Write(item.Value); err != nil {
        log.Println("unable to write image.")
    }	
}

func beard(c appengine.Context) image.Image {
	item, err := memcache.Get(c, "beard")
	
	if err == memcache.ErrCacheMiss {
	    c.Infof("beard not in the cache")
			
	    client := urlfetch.Client(c)
	    resp, err := client.Get("http://greenyfy.appspot.com/images/beard.png")
	
	    if err != nil {
	        log.Println("Failed to get beard url")
	    }
	    
	    defer resp.Body.Close()
	    
		buff := new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		
		item = &memcache.Item{
		    Key:   "beard",
		    Value: buff.Bytes(),
		}
		
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		    c.Infof("item with key %q already exists", item.Key)
		} else if err != nil {
		    c.Errorf("error adding item: %v", err)
		}
	} else if err != nil {
	    c.Errorf("error getting item: %v", err)
	} else {
		log.Println("Cache hit beard")
	}
	
	c.Infof("found beard in memcache")
	
	buff := bytes.NewReader(item.Value)
	
	img, _, _ := image.Decode(buff)
	
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