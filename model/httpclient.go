package model

import (
	"fmt"
	"github.com/gregjones/httpcache"
	httpdc "github.com/gregjones/httpcache/diskcache"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// NewHTTPClient uses the settings to create a new, thread-safe and reusable, HTTP Client.
// The resulting http.Client will always be a valid, even if there's an error.
func (hcs HTTPClientSettings) NewHTTPClient() (*http.Client, error) {
	if hcs.Cache == nil {
		return hcs.newDefaultHTTPClient(), nil
	}

	switch cache := hcs.Cache.(type) {
	case *HTTPDiskCache:
		return hcs.newHTTPDiskCacheClient(cache)
	case *HTTPMemoryCache:
		return hcs.newHTTPMemoryCacheClient(cache)
	default:
		return hcs.newDefaultHTTPClient(), fmt.Errorf("Unknown cache type %T in HTTPClientSettings.NewHTTPClient()", hcs.Cache)
	}
}

func (hcs HTTPClientSettings) newDefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: time.Duration(hcs.Timeout)}
}

func (hcs HTTPClientSettings) newHTTPMemoryCacheClient(cache *HTTPMemoryCache) (*http.Client, error) {
	fmt.Println("Using built-in httpcache.MemoryCache for HTTPClient cache")
	return &http.Client{Transport: httpcache.NewTransport(httpcache.NewMemoryCache()), Timeout: time.Duration(hcs.Timeout)}, nil
}

func (hcs HTTPClientSettings) newHTTPDiskCacheClient(cache *HTTPDiskCache) (*http.Client, error) {
	cacheDir, err := filepath.Abs(cache.BasePath)
	if err != nil {
		cache.Activities.AddError("HTTPClientSettings", "HHTTPC-0001", fmt.Sprintf("%q is not a valid BasePath: %s", cache.BasePath, err.Error()))
		return hcs.newDefaultHTTPClient(), err
	}

	if _, err = os.Stat(cacheDir); os.IsNotExist(err) {
		if cache.CreateBasePath {
			err = os.MkdirAll(cacheDir, os.FileMode(0755))
			if err != nil {
				cache.Activities.AddError("HTTPClientSettings", "HHTTPC-0002", fmt.Sprintf("Unable to create BasePath %q: %s", cacheDir, err.Error()))
				return &http.Client{Timeout: time.Duration(hcs.Timeout)}, err
			}
			cache.Activities.AddHistory(&ActivityLog{Message: ActivityHumanMessage(fmt.Sprintf("Created HTTPDiskCache with BasePath %q", cacheDir))})
		}
	} else {
		cache.Activities.AddHistory(&ActivityLog{Message: ActivityHumanMessage(fmt.Sprintf("Using existing HTTPDiskCache BasePath %q", cacheDir))})
	}

	httpCache := httpdc.New(cacheDir)
	cache.Activities.AddHistory(&ActivityLog{Message: ActivityHumanMessage(fmt.Sprintf("Using %+v for HTTP Disk Cache", httpCache))})

	for _, a := range cache.Activities.History {
		switch log := a.(type) {
		case *ActivityLog:
			fmt.Println(log.Message)
		}
	}

	return &http.Client{Transport: httpcache.NewTransport(httpCache), Timeout: time.Duration(hcs.Timeout)}, nil
}
