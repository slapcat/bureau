package main

import (
	"os"
	"log"
	"reflect"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug       bool   `default:"false"`
	Daemon      bool   `default:"true"`
	Server      string
	Binddn      string
	Password    string
	Base        string
	Update      int    `yaml:"update_interval" default:"600"`
	Host        bool   `yaml:"host_specific_entries" default:"true"`
	Restart     bool   `yaml:"restart_service_on_change" default:"true"`
	Override    string `yaml:"override_hostname"`
	HostDN      string
  Local2LDAP  bool   `yaml:"sync_local_changes" default:"false"`
}

func ConfigInit() (Config, error) {

	var conf string

	home, err := os.UserHomeDir()
	locations := []string{
		home + "/.bureau.yaml",
		home + "/.config/bureau/bureau.yaml",
		"/etc/bureau/bureau.yaml",
		"bureau.yaml",
	}

	for _, path := range locations {
		if _, err = os.Stat(path); err == nil {
			conf = path
			break
		}
	}

	data, err := os.ReadFile(conf)
	if err != nil {
		return Config{}, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatalf("Error: %v", err)
	} else if c.Debug {
		for i := 0; i <= 9; i++ {
			key := reflect.Indirect(reflect.ValueOf(c)).Type().Field(i).Name
			value := reflect.ValueOf(c)
			if key == "Password" && c.Password != "" {
				log.Printf(" === Loading configuration %v: %s\n", key, "***HIDDEN PASSWORD***")
			} else {
				log.Printf(" === Loading configuration %v: %v\n", key, value.FieldByName(key))
			}
		}
	}

	// get hostname and set search base
	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("error: %v", err)
	} else if c.Debug {
		log.Printf(" === Getting hostname: %s", host)
	}

	if c.Host {
		if c.Override != "" {
			c.HostDN = "cn=" + c.Override + "," + c.Base
		} else {
			c.HostDN = "cn=" + host + "," + c.Base
		}
	} else {
		c.HostDN = c.Base
	}
	if c.Debug {
		log.Printf(" === Looking for files in: %s", c.HostDN)
	}

	return c, nil
}
