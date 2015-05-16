package greenyfy

import (
    "bytes"
    "fmt"
    "log"
    "net/http"
    "image"
    "image/jpeg"
    _ "image/png"
    "strconv"
    "github.com/nfnt/resize"
    
    "appengine"
    "appengine/urlfetch"
)

func init() {
    http.HandleFunc("/me", me)
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Because API")
}

func me(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    client := urlfetch.Client(c)
    resp, err := client.Get("http://images2.fanpop.com/image/photos/9200000/Pretty-Odd-pretty-odd-photography-9283045-1600-1200.jpg")
    
    defer resp.Body.Close()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
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
    
    
    writeImage(w, &img)
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