package main

import (
//	"fmt"
	"reflect"
	"os"
    "log"
	"time"
    "gopkg.in/yaml.v3"
)


type Config struct {
	Debug bool
	Binddn string
	Password string
	Base string
	Update int `yaml:"update_interval"`
	Host bool `yaml:"host_specific_entries"`
	Restart bool `yaml:"restart_service_on_change"`
	Files []string
}

func main() {
	data, err := os.ReadFile("config.yaml")
    if err != nil {
		log.Fatalf("error: %v", err)
    }

    c := Config{}
    err = yaml.Unmarshal([]byte(data), &c)
    if err != nil {
		log.Fatalf("error: %v", err)
    } else if c.Debug == true {
		for i := 0; i <= 7; i++ {
			key := reflect.Indirect(reflect.ValueOf(c)).Type().Field(i).Name
			value := reflect.ValueOf(c)
			    log.Printf(" === Loading configuration %v: %v\n", key, value.FieldByName(key))
		}
	}

	for {
	// ldap bind

	// compare lastmodified date, load config files: based on c.files

	// call appropriate generator to generate file
    
	log.Print("test")
	time.Sleep(time.Duration(c.Update) * time.Second)
	// sleep
	}	
}
