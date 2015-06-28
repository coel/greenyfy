package greenyfy

import (
    "bytes"
    
    "image"
    _ "image/jpeg"
    _ "image/png"

    "appengine"
    "appengine/urlfetch"
)

var beardCacheKey = "beard"

func getBeardCached(c appengine.Context) (image.Image, error) {
    
    item, err := getCached(c, beardCacheKey, getBeardFromUrl)
    
    if err != nil {
        return nil, err
    }
            
    buff := bytes.NewReader(item.Value)
    
    img, _, err := image.Decode(buff)
    
    if err != nil {
        return nil, err
    }
    
    return img, nil
}

func getBeardFromUrl (c appengine.Context, key string) (*bytes.Buffer, error) {
    client := urlfetch.Client(c)
    resp, err := client.Get("http://" + appengine.DefaultVersionHostname(c) + "/images/beard.png")

    if err != nil {
        return nil, err
    }
    
    defer resp.Body.Close()
    
    buff := new(bytes.Buffer)
    buff.ReadFrom(resp.Body)
    
    return buff, nil
}