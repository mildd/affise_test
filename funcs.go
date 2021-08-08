package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func OnErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func parseBody(w http.ResponseWriter, r *http.Request) ([]string, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Body read error, %v", err)
		w.WriteHeader(500)
		return nil, err
	}
	defer r.Body.Close()
	var urls []string
	if err = json.Unmarshal(body, &urls); err != nil {
		log.Printf("Body unmarshal error, %v", err)
		w.WriteHeader(400)
		return nil, err
	}
	return urls, nil
}

func makeRequests(urls []string) (map[string]string, error) {
	respMap := make(map[string]string)
	time.Sleep(15 * time.Second)
	for _, url := range urls {
		ctx := context.Background()
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))
		defer cancel()
		resp, err := doRequest(ctx, url)
		if resp == nil {
			continue
		}
		if err != nil {
			log.Printf("Error with Get request, %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Body read error, %v", err)
				return nil, err
			}
			respMap[url] = string(body)
		} else {
			err = fmt.Errorf("%s Responded with status code - %d", url, resp.StatusCode)
			return nil, err
		}
	}
	return respMap, nil
}

func doRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
	}
	req = req.WithContext(ctx)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	select {
	case <-time.After(500 * time.Millisecond):
		fmt.Println("request completed")
		return res, err
	case <-ctx.Done():
		fmt.Println("request too long")
		return nil, nil
	}
}

func maxClientsMiddleware(h http.Handler, n int) http.Handler {
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()

		h.ServeHTTP(w, r)
	})
}
