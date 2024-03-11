package main

import (
//	"fmt"
	"reflect"
	"os"
        "log"
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
}

func main() {
	data, err := os.ReadFile("config.yaml")
        if err != nil {
                log.Fatalf("error: %v", err)
        }

        c := Config{}
        err2 := yaml.Unmarshal([]byte(data), &c)
        if err2 != nil {
                log.Fatalf("error: %v", err)
        }

	if c.Debug == true {
		for i := 0; i <= 6; i++ {
			key := reflect.Indirect(reflect.ValueOf(c)).Type().Field(i).Name
			value := reflect.ValueOf(c)
		        log.Printf(" === Loading configuration %v: %v\n", key, value.FieldByName(key))
		}
	}

}
