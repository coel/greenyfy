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
	"appengine/memcache"
    
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
	
	item, err := memcache.Get(c, img_url)
	
	if err == memcache.ErrCacheMiss {
	    c.Infof("item not in the cache")
		
	    client := urlfetch.Client(c)
	    resp, err := client.Get(img_url)
	
	    if err != nil {
	        c.Infof("Failed to get url: ", img_url)
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
	        c.Infof("Resizing image", bnds.Dx())
	        img = resize.Resize(1024, 0, img, resize.Lanczos3)
	    }
	    
		faces := findFaces(c, &img)
		// todo: should I pass back by reference?
		
		c.Infof("Obj: ", faces)
			
	    bnds = img.Bounds()
	
	    m := image.NewRGBA(image.Rect(0, 0, bnds.Dx(), bnds.Dy()))
	    
	    draw.Draw(m, bnds, img, image.Point{0,0}, draw.Src)
	    
	    brd, err := getBeard(c)
		if (err != nil) {
	        http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	        c.Infof("unable to encode image.")
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
		c.Infof("Cache hit: ", img_url)
	}

    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(len(item.Value)))
    if _, err := w.Write(item.Value); err != nil {
        c.Infof("unable to write image.")
    }	
}
