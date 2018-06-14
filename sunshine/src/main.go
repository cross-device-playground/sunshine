package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type mobile_sensing_data struct {
	ID              string  `json:"ID"`
	DeviceName      string  `json:"DeviceName"`
	Timestamp       int64   `json:"Timestamp"`
	Compass         float32 `json:"Compass"`
	Accelerometer_x float32 `json:"Accelerometer_x"`
	Accelerometer_y float32 `json:"Accelerometer_y"`
	Accelerometer_z float32 `json:"Accelerometer_z"`
	Gyroscope_x     float32 `json:"Gyroscope_x"`
	Gyroscope_y     float32 `json:"Gyroscope_y"`
	Gyroscope_z     float32 `json:"Gyroscope_z"`
	Magnetometer_x  float32 `json:"Magnetometer_x"`
	Magnetometer_y  float32 `json:"Magnetometer_y"`
	Magnetometer_z  float32 `json:"Magnetometer_z"`
}

var mobile_sensing_data_csv_col = []string{"ID", "DeviceName", "Timestamp", "Compass", "Accelerometer_x", "Accelerometer_y", "Accelerometer_z", "Gyroscope_x", "Gyroscope_y", "Gyroscope_z", "Magnetometer_x", "Magnetometer_y", "Magnetometer_z"}

func mobile_sensing_data_to_str(data mobile_sensing_data) []string {
	return []string{data.ID, data.DeviceName, strconv.FormatInt(data.Timestamp, 10), fmt.Sprintf("%.6f", data.Compass), fmt.Sprintf("%.6f", data.Accelerometer_x), fmt.Sprintf("%.6f", data.Accelerometer_y), fmt.Sprintf("%.6f", data.Accelerometer_z), fmt.Sprintf("%.6f", data.Gyroscope_x), fmt.Sprintf("%.6f", data.Gyroscope_y), fmt.Sprintf("%.6f", data.Gyroscope_z), fmt.Sprintf("%.6f", data.Magnetometer_x), fmt.Sprintf("%.6f", data.Magnetometer_y), fmt.Sprintf("%.6f", data.Magnetometer_z)}
}

var m sync.Mutex

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

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			// fmt.Fprintf(w, "404")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello")
	})
}

var num = 0

func mobile_sensing(logger *log.Logger, writer *csv.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var data mobile_sensing_data
		err := decoder.Decode(&data)
		if err != nil {
			panic(err)
		}
		m.Lock()
		defer m.Unlock()
		err = writer.Write(mobile_sensing_data_to_str(data))
		if err != nil {
			logger.Println("here!!!!", err)
		}
		writer.Flush()
		num += 1
		logger.Println("mobile", num, "-", data)
		logger.Println(mobile_sensing_data_to_str(data))
	})
}

func main() {

	// Data .csv file
	file, err := os.Create("data/mobile.csv")
	if err != nil {
		log.Fatal("Cannot create mobile.csv", err)
	}

	writer := csv.NewWriter(file)
	err = writer.Write(mobile_sensing_data_csv_col)
	if err != nil {
		log.Fatal("Cannot write to mobile.csv ")
	}
	writer.Flush()
	defer file.Close()

	// Cross Device Playground Server
	logger := log.New(os.Stdout, "Cross Device Playground: ", log.LstdFlags)
	logger.Println("Cross-Device Playground (CDP) Server v0.1")
	logger.Println("CDP Server starting...")

	router := http.NewServeMux()
	router.Handle("/", index())
	router.Handle("/mobile/sensing/", mobile_sensing(logger, writer))

	var ip_address = get_ip_address()
	var port_number = 5566
	var server_address = fmt.Sprintf("%s:%d", ip_address, port_number)

	server := &http.Server{
		Addr:         server_address,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("CDP Server shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Printf("CDP Server started at %s !", server_address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", server_address, err)
	}

	<-done
	logger.Println("CDP Server stopped !")
}
