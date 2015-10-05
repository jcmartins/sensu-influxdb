package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ListenHost       string `json:"listen_host"`
	ListenPort       string `json:"listen_port"`
	ListenConnType   string `json:"listen_conn_type"`
	InfluxDbHost     string `json:"influxdb_host"`
	InfluxDbPort     string `json:"influxdb_port"`
	InfluxDbUsername string `json:"influxdb_username"`
	InfluxDbPassword string `json:"influxdb_password"`
	InfluxDbDatabase string `json:"influxdb_database"`
	LogReceived      bool   `json:"log_received"`
}

func getConfig() (Config, error) {
	c := Config{
		ListenHost:       "0.0.0.0",
		ListenPort:       "3333",
		ListenConnType:   "tcp",
		InfluxDbHost:     "127.0.0.1",
		InfluxDbPort:     "8086",
		InfluxDbUsername: "",
		InfluxDbPassword: "",
		InfluxDbDatabase: "sensu",
		LogReceived:      false,
	}
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		return Config{}, err
	}
	err = json.Unmarshal(raw, &c)
	return c, err
}
