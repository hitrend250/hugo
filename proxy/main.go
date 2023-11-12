package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Address struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Value             string `json:"value"`
	UnrestrictedValue string `json:"unrestricted_value"`
}

type RequestAddressSearch struct {
	Query string `json:"query"`
}

//type ResponseAddress struct {
//	Addresses []*Address `json:"addresses"`
//}
//
//type RequestAddressGeocode struct {
//	Lat string `json:"lat"`
//	Lng string `json:"lng"`
//}

func main() {

	r := chi.NewRouter()

	proxy := NewReverseProxy("hugo", "1313")
	r.Use(proxy.ReverseProxy)
	r.Get("/api/", apiHelloHandler)
	r.Group(func(r chi.Router) {
		r.Post("/api/address/search", apiSearchHandler)
	})
	http.ListenAndServe(":8080", r)
}
func apiHelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from API"))
}
func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	var re RequestAddressSearch
	err := json.NewDecoder(r.Body).Decode(&re)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	jsonData, err := json.Marshal(re)
	var data = strings.NewReader(string(jsonData))
	fmt.Println(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token 864ecfb76388cdeb4ee1f7215e1eb8272f5d56b7")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	var dadataResponse Address

	err = json.NewDecoder(resp.Body).Decode(&dadataResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(dadataResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ReverseProxy struct {
	host string
	port string
}

func NewReverseProxy(host, port string) *ReverseProxy {
	return &ReverseProxy{
		host: host,
		port: port,
	}
}

func (rp *ReverseProxy) ReverseProxy(next http.Handler) http.Handler {
	target, _ := url.Parse(fmt.Sprintf("http://%s:%s", rp.host, rp.port))
	proxy := httputil.NewSingleHostReverseProxy(target)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/address/search" || r.URL.Path == "/api/address/geocode" {
			next.ServeHTTP(w, r)
		} else {
			proxy.ServeHTTP(w, r)
		}
	})
}
