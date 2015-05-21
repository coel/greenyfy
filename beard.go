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



func getBeard(c appengine.Context) (image.Image, error) {
	
	item, err := getCached(c, beardCacheKey, getBeardFromUrl)
		
	buff := bytes.NewReader(item.Value)
	
	img, _, err := image.Decode(buff)
	
	if err != nil {
	    return nil, err
	}
	
    return img, nil
}

func getBeardFromUrl (c appengine.Context, key string) (*bytes.Buffer, error) {
    client := urlfetch.Client(c)
    resp, err := client.Get(config.BeardUrl)

    if err != nil {
        return nil, err
    }
    
    defer resp.Body.Close()
    
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	
	return buff, nil
}