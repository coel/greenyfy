package greenyfy

import (
    "bytes"
    "fmt"
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
	
	item, err := getCached(c, img_url, do)
	
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(item.Value)))
    if _, err := w.Write(item.Value); err != nil {
        c.Infof("unable to write image.")
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }	
}

func do (c appengine.Context, key string) (*bytes.Buffer, error) {

    client := urlfetch.Client(c)
    resp, err := client.Get(key)

    if err != nil {
        return nil, err
    }
    
    defer resp.Body.Close()
    
    // fmt.Fprintf(w, "HTTP GET returned status %v", resp.Status)
    
    img, _, err := image.Decode(resp.Body)
    if err != nil {
        return nil, err
    }
    
    bnds := img.Bounds()
    if bnds.Dx() > 1024 {
        c.Infof("Resizing image", bnds.Dx())
        img = resize.Resize(1024, 0, img, resize.Lanczos3)
    }
    
	faces, err := findFaces(c, &img)
	// todo: should I pass back by reference?
	if err != nil {
        return nil, err
    }
	
	c.Infof("Obj: ", faces)
		
    bnds = img.Bounds()

    m := image.NewRGBA(image.Rect(0, 0, bnds.Dx(), bnds.Dy()))
    
    draw.Draw(m, bnds, img, image.Point{0,0}, draw.Src)
    
    brd, err := getBeard(c)
	if (err != nil) {
		return nil, err
	}
	
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
	
    img_out := image.Image(m)
	    
		
    buffer := new(bytes.Buffer)
    if err := jpeg.Encode(buffer, img_out, nil); err != nil {
        return nil, err
    }

	return buffer, nil
}