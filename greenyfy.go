package greenyfy

import (
    "bytes"
    "fmt"
    "log"
    "net/http"
    "image"
    "image/jpeg"
    "strconv"
)

func init() {
    http.HandleFunc("/me", me)
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Because API")
}

func me(w http.ResponseWriter, r *http.Request) {
    
    resp, err := http.Get("http://golang.org/doc/gopher/frontpage.png")
    if err != nil {
        log.Fatal(err)
    }
    _, format, _ := image.Decode(resp.Body)
    defer resp.Body.Close()
    fmt.Fprint(w, "me", format)
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