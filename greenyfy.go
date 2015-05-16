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
    
    bnds = img.Bounds()
    
    brd := beard(c)
    //log.Println("beard: ", brd)
    
    sr := image.Rect(110,110,500,550)
    dp := image.Point{10, 10}
    rt := image.Rectangle{dp, dp.Add(sr.Size())}
    m := image.NewRGBA(image.Rect(0, 0, bnds.Max.X, bnds.Max.Y))
    
    draw.Draw(m, bnds, img, image.Point{0,0}, draw.Src)
    draw.Draw(m, rt, brd, sr.Min, draw.Over)
    
    var img_out image.Image = m
    
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