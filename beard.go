package greenyfy

import (
    "bytes"
    
    "image"
    _ "image/jpeg"
    _ "image/png"

    "appengine"
    "appengine/urlfetch"
	"appengine/memcache"
)

func getBeard(c appengine.Context) (image.Image, error) {
	item, err := memcache.Get(c, "beard")
	
	if err == memcache.ErrCacheMiss {
	    c.Infof("beard not in the cache")
			
	    client := urlfetch.Client(c)
	    resp, err := client.Get(config.BeardUrl)
	
	    if err != nil {
	        return nil, err
	    }
	    
	    defer resp.Body.Close()
	    
		buff := new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		
		item = &memcache.Item{
		    Key:   "beard",
		    Value: buff.Bytes(),
		}
		
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		    c.Warningf("item with key %q already exists", item.Key)
		} else if err != nil {
		    return nil, err
		}
	} else if err != nil {
	    return nil, err
	}
	
	c.Infof("found beard in memcache")
	
	buff := bytes.NewReader(item.Value)
	
	img, _, _ := image.Decode(buff)
	
    return img, nil
}
