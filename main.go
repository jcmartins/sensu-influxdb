package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"

	"github.com/influxdb/influxdb/client"

	"strconv"
	"strings"
	"time"
)

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal("Error obtaining the config", err)
	}
	log.Println("config", config)

	l, err := net.Listen(config.ListenConnType, config.ListenHost+":"+config.ListenPort)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer l.Close()

	log.Println("Listening on " + config.ListenHost + ":" + config.ListenPort)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
		} else {
			go handleRequest(conn, config)
		}
	}
}

type Client struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Check struct {
	Output string `json:"output"`
}

type Event struct {
	Client Client `json:"client"`
	Check  Check  `json:"check"`
}

func getInfluxdbClient(config Config) (*client.Client, error) {
	host, err := url.Parse(fmt.Sprintf("http://%s:%s", config.InfluxDbHost, config.InfluxDbPort))
	if err != nil {
		return nil, err
	}
	conf := client.Config{
		URL:      *host,
		Username: config.InfluxDbUsername,
		Password: config.InfluxDbPassword,
	}
	c, err := client.NewClient(conf)
	if err != nil {
		return nil, err
	}
	_, _, err = c.Ping()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func handleRequest(conn net.Conn, config Config) {
	c, err := getInfluxdbClient(config)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()
	buf, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Println("Error reading:", err.Error())
		return
	}

	if config.LogReceived == true {
		log.Println(string(buf))
	}

	var evt Event
	err = json.Unmarshal(buf, &evt)
	if err != nil {
		log.Printf("Error unmarshalling event: %s for input %s\n", err.Error(), string(buf))
		return
	}

	outputlines := strings.Split(strings.TrimSpace(evt.Check.Output), "\n")

	pts := make([]client.Point, len(outputlines))

	for k, l := range outputlines {
		line := strings.TrimSpace(l)
		pieces := strings.Split(line, " ")
		if len(pieces) != 3 {
			log.Println("Wrong number of pieces")
			continue
		}
		keys := strings.SplitN(pieces[0], ".", 2)
		if len(keys) != 2 {
			log.Println("Wrong number of pieces")
			continue
		}
		keyraw := keys[1]
		key := strings.Replace(keyraw, ".", "_", -1)

		val, verr := strconv.ParseFloat(pieces[1], 64)
		if verr != nil {
			log.Printf("Error parsing value (%s): %s\n", pieces[1], verr.Error())
			continue
		}

		_, terr := strconv.ParseInt(pieces[2], 10, 64)
		if terr != nil {
			log.Printf("Error parsing time (%s): %s\n", pieces[2], terr.Error())
			continue
		}

		pts[k] = client.Point{
			Measurement: key,
			Tags: map[string]string{
				"client_name":    evt.Client.Name,
				"client_address": evt.Client.Address,
			},
			Fields: map[string]interface{}{
				"value": float32(val),
			},
			Time: time.Now(),
		}
		log.Println("val", val, float32(val))
	}

	bps := client.BatchPoints{
		Points:          pts,
		Database:        config.InfluxDbDatabase,
		RetentionPolicy: "default",
	}

	r, err := c.Write(bps)
	if config.LogReceived == true {
		log.Printf("%#v", r)
	}
	if err != nil {
		log.Println("write error:", err)
	}

}
