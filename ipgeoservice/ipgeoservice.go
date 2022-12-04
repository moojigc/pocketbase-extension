package ipgeoservice

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type Connection struct {
	Isp string `json:"isp_name"`
	Org string `json:"organization_name"`
}

type IpGeoService struct {
	Ip         string     `json:"ip_address"`
	Country    string     `json:"country"`
	Region     string     `json:"region"`
	City       string     `json:"city"`
	PostalCode string     `json:"postal_code"`
	Latitude   float64    `json:"latitude"`
	Longitude  float64    `json:"longitude"`
	Connection Connection `json:"connection"`
}

var apiKey = os.Getenv("ABSTRACT_API_KEY")

var logger *log.Logger = log.Default()

func GetIpGeo(ip string) *IpGeoService {

	resp, err := http.Get("https://ipgeolocation.abstractapi.com/v1?ip_address=" + ip + "&api_key=" + apiKey)

	if err != nil || resp.StatusCode != 200 {
		logger.Printf("Error getting ip geo for %s: %s", ip, err)
	}

	body, err := io.ReadAll(resp.Body)

	logger.Print(string(body))

	if err != nil {
		logger.Println(err)
	}

	defer resp.Body.Close()

	var ipGeo IpGeoService

	if err := json.Unmarshal(body, &ipGeo); err != nil {
		panic(err)
	}

	return &ipGeo
}
