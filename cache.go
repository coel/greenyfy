package greenyfy

import (
    "bytes"
    "appengine"
	"appengine/memcache"
)

func getCached(c appengine.Context, key string, missing func(appengine.Context, string) (*bytes.Buffer, error)) (*memcache.Item, error) {

	item, err := memcache.Get(c, key)
	
	if err == memcache.ErrCacheMiss {
	    c.Infof("item not in the cache: ", key)
		
		result, _ := missing(c, key)
		
		item = &memcache.Item{
		    Key:   key,
		    Value: result.Bytes(),
		}
		
		if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		    c.Warningf("item with key %q already exists", item.Key)
		} else if err != nil {
		    return item, err
		}
	} else if err != nil {
	    return item, err
	} else {
		c.Infof("Cache hit: ", key)
	}
	
	return item, nil
}