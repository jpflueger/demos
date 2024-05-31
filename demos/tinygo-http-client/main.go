package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/variables"
)

type ApiObj struct {
	Id        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := spinhttp.NewRouter()
		router.GET("/", listObjects)
		router.PUT("/:id", updateObject)
		router.POST("/", createObject)
		router.ServeHTTP(w, r)
	})
}

func main() {}

func listObjects(w http.ResponseWriter, r *http.Request, p spinhttp.Params) {
	fmt.Println("1")
	apiReq, err := http.NewRequest("GET", "https://api.restful-api.dev/objects", bytes.NewReader(make([]byte, 0)))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create outbound http request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	fmt.Println("2")
	apiRes, err := mySend(apiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to execute outbound http request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if apiRes.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("response from outbound http request is not OK %v", apiRes.Status), http.StatusInternalServerError)
		return
	}

	fmt.Println("3")
	apiBody, err := io.ReadAll(apiRes.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read outbound http response: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	fmt.Println("4")
	if len(apiBody) == 0 {
		http.Error(w, fmt.Sprintf("outbound http response was empty\n"), http.StatusInternalServerError)
		return
	}

	fmt.Println("5")
	apiObjs := make([]ApiObj, 0)
	err = json.Unmarshal(apiBody, &apiObjs)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode outbound http response from json: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	fmt.Println("6")
	w.Header().Set("Content-Type", "application/json")
	we := json.NewEncoder(w)
	we.SetIndent("", "")
	we.Encode(apiObjs)
}

func createObject(w http.ResponseWriter, r *http.Request, p spinhttp.Params) {
	if r.ContentLength == 0 {
		http.Error(w, "request body must contain a json object", http.StatusBadRequest)
		return
	}

	var obj ApiObj
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		http.Error(w, fmt.Sprintf("request body must contain a json object: %v", err.Error()), http.StatusBadRequest)
		return
	}

	objBody, err := json.Marshal(&obj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode object as json: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	postReq, err := http.NewRequest("POST", "https://api.restful-api.dev/objects", bytes.NewBuffer(objBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create outgoing http post request: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	postRes, err := mySend(postReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to send outgoing http post request: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	if postRes.StatusCode != 200 {
		http.Error(w, fmt.Sprintf("outgoing http post response failed: %v", postRes.Status), postRes.StatusCode)
		return
	}

	n, err := io.Copy(w, postRes.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("outgoing http post response failed: %v", postRes.Status), http.StatusInternalServerError)
		return
	}

	if n != postRes.ContentLength {
		fmt.Println("wrote", n, "bytes but expected", postRes.ContentLength)
	}
}

func updateObject(w http.ResponseWriter, r *http.Request, p spinhttp.Params) {
	if r.ContentLength == 0 {
		http.Error(w, "request body must contain a json object", http.StatusBadRequest)
		return
	}

	id := p.ByName("id")
	if id == "" {
		http.Error(w, "request url must contain an identifier", http.StatusBadRequest)
		return
	}

	var obj ApiObj
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		http.Error(w, fmt.Sprintf("request body must contain a json object: %v", err.Error()), http.StatusBadRequest)
		return
	}

	objBody, err := json.Marshal(&obj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode object as json: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	putReq, err := http.NewRequest("PUT", fmt.Sprintf("https://api.restful-api.dev/objects/%s", id), bytes.NewBuffer(objBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create outgoing http put request: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	putRes, err := mySend(putReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to send outgoing http put request: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	if putRes.StatusCode != 200 {
		http.Error(w, fmt.Sprintf("outgoing http put response failed: %v", putRes.Status), putRes.StatusCode)
		return
	}

	n, err := io.Copy(w, putRes.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("outgoing http put response failed: %v", putRes.Status), http.StatusInternalServerError)
		return
	}

	if n != putRes.ContentLength {
		fmt.Println("wrote", n, "bytes but expected", putRes.ContentLength)
	}
}

func mySend(req *http.Request) (*http.Response, error) {
	req.Header.Set("content-type", "application/json")

	sender, _ := variables.Get("sender")
	switch sender {
	case "http.DefaultClient.Do":
		return http.DefaultClient.Do(req)
	case "":
		fallthrough
	case "spinhttp.Send":
		fallthrough
	default:
		return spinhttp.Send(req)
	}
}
