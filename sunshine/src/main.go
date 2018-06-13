package main

import (
	"fmt"
	"log"
	"os"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type mobile_sensing_data struct {
	ID string `json:"ID"`
	DeviceName string `json:"DeviceName"`
	Timestamp int64 `json:"Timestamp"`
	Compass float32 `json:"Compass"`
	Accelerometer_x float32 `json:"Accelerometer_x"`
	Accelerometer_y float32 `json:"Accelerometer_y"`
	Accelerometer_z float32 `json:"Accelerometer_z"`
	Gyroscope_x float32 `json:"Gyroscope_x"`
	Gyroscope_y float32 `json:"Gyroscope_y"`
	Gyroscope_z float32 `json:"Gyroscope_z"`
	Magnetometer_x float32 `json:"Magnetometer_x"`
	Magnetometer_y float32 `json:"Magnetometer_y"`
	Magnetometer_z float32 `json:"Magnetometer_z"`
} 

func main() {
	file, err := os.Create("data/mobile.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fmt.Fprintf(w, "404")
			return
		}
		fmt.Fprintf(w, "hello")
		log.Println("hello")
	})

	http.HandleFunc("/mobile/sensing/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var data mobile_sensing_data
		err := decoder.Decode(&data)
		if err != nil {
			panic(err)
		}
		log.Println("mobile |", data)
	})

	var ip_address = get_ip_address()
	var port_number = 5566
	var host = fmt.Sprintf("%s:%d", ip_address, port_number)

	log.Println("Starting Cross-Device Playground (CDP) server @", host)
	log.Fatal(http.ListenAndServe(host, nil))
}

func get_ip_address() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}