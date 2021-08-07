package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error with Get request, %v", err)
			return nil, err
		}
		// TODO !?
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Body read error, %v", err)
				return nil, err
			}
			respMap[url] = string(body)
		} else {
			err = fmt.Errorf("%s responded with status code - %d", url, resp.StatusCode)
			return nil, err
		}
	}
	return respMap, nil
}

func main() {
	h1 := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			urls, err := parseBody(w, r)
			if err != nil {
				return
			}
			if len(urls) > 20 {
				log.Println("Too many urls in request")
				w.WriteHeader(400)
				return
			}
			respMap := make(map[string]string)
			respMap, err = makeRequests(urls)
			if err != nil {
				w.WriteHeader(500)
				_, err = w.Write([]byte(fmt.Sprintf("Error: %s", err)))
				OnErr(err)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(respMap)
			if err != nil {
				log.Printf("Marshaling error, %v", err)
				w.WriteHeader(500)
				return
			}
			_, err = w.Write(obj)
			if err != nil {
				log.Printf("Error while writing in ResponseWriter: %v", err)
				w.WriteHeader(500)
				return
			}
		} else {
			fmt.Println("Method not allowed")
			w.WriteHeader(405)
		}
	}

	http.HandleFunc("/get_info/", h1)

	srv := new(Server)
	log.Fatal(srv.Run("8080"))
}
