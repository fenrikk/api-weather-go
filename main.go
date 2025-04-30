package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type IPLocation struct {
	Status  string  `json:"status"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	City    string  `json:"city"`
	Country string  `json:"country"`
}

type WeatherResponse struct {
	Metadata    interface{} `json:"metadata"`
	Units       interface{} `json:"units"`
	DataCurrent interface{} `json:"data_current"`
}

type Config struct {
	MeteoBlueAPIKey string
	Port            string
}

func main() {
	apiKey := os.Getenv("METEOBLUE_API_KEY")
	if apiKey == "" {
		log.Fatal("Environment variable METEOBLUE_API_KEY is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	config := Config{
		MeteoBlueAPIKey: apiKey,
		Port:            port,
	}

	http.HandleFunc("/getWeather", handleGetWeather(config))

	fmt.Printf("Server started on port %s...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func handleGetWeather(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ip string
		
		ipParam := r.URL.Query().Get("ip")
		if ipParam != "" {
			ip = ipParam
			fmt.Printf("Using IP from request parameter: %s\n", ip)
		} else {
			ip = getClientIP(r)
			fmt.Printf("Detected client IP: %s\n", ip)
		}

		location, err := getLocationFromIP(ip)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting location: %v", err), http.StatusInternalServerError)
			return
		}

		weatherData, err := getWeatherData(location.Lat, location.Lon, config.MeteoBlueAPIKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting weather data: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(weatherData)
	}
}

func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func getLocationFromIP(ip string) (*IPLocation, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query IP location API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IP location API response: %w", err)
	}

	var location IPLocation
	if err := json.Unmarshal(body, &location); err != nil {
		return nil, fmt.Errorf("failed to parse IP location API response: %w", err)
	}

	if location.Status != "success" {
		return nil, fmt.Errorf("IP location API returned non-success status: %s", location.Status)
	}

	return &location, nil
}

func getWeatherData(lat, lon float64, apiKey string) (*WeatherResponse, error) {
	url := fmt.Sprintf("https://my.meteoblue.com/packages/current?apikey=%s&lat=%f&lon=%f&format=json", 
		apiKey, lat, lon)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query weather API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read weather API response: %w", err)
	}

	var weatherData WeatherResponse
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return nil, fmt.Errorf("failed to parse weather API response: %w", err)
	}

	return &weatherData, nil
} 