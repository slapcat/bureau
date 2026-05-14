package main

import (
	"os"
	"fmt"
	"reflect"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug       bool
	Daemon      bool
	Server      string
	Binddn      string
	Password    string
	Base        string
	Update      int    `yaml:"update_interval"`
	Host        bool   `yaml:"host_specific_entries"`
	Restart     bool   `yaml:"restart_service_on_change"`
	Override    string `yaml:"override_hostname"`
	HostDN      string
	Local2LDAP  bool   `yaml:"sync_local_changes"`
}

func ConfigInit() (Config, error) {

	var conf string
	var data []byte

	home, err := os.UserHomeDir()
	locations := []string{
		"bureau.yaml",
		home + "/.bureau.yaml",
		home + "/.config/bureau/bureau.yaml",
		"/etc/bureau/bureau.yaml",
	}

	for _, path := range locations {
		if _, err = os.Stat(path); err == nil {
			conf = path
			break
		}
	}
	
	if conf != "" {
		data, err = os.ReadFile(conf)
		
		err = yaml.Unmarshal(data, &C)
		Logger(err, "Unmarshal error", "FATAL")
	} else {
		C = Config{
				Debug: true,
				Daemon: true,
				Server: "ldap://opendirectory.net",
				Binddn: "cn=example,ou=admins,dc=opendirectory,dc=net",
				Password: "Passw0rd",
				Base: "ou=config,dc=example,dc=opendirectory,dc=net",
				Update: 600,
				Host: true,
				Restart: true,
				Override: "bureau",
			}
	}
	
	if C.Debug {
		for i := 0; i <= 9; i++ {
			key := reflect.Indirect(reflect.ValueOf(C)).Type().Field(i).Name
			value := reflect.ValueOf(C)

			if key == "Password" && C.Password != "" {
				Logger(nil, fmt.Sprintf("Loading configuration %v: %s\n", key, "***HIDDEN PASSWORD***"), "DEBUG")
			} else {
				Logger(nil, fmt.Sprintf("Loading configuration %v: %v\n", key, value.FieldByName(key)), "DEBUG")
			}
		}
	}

	// get hostname and set search base
	host, err := os.Hostname()
	Logger(err, "Error", "FATAL")
	Logger(nil, "Getting hostname: "+host, "DEBUG")

	if C.Host {
		if C.Override != "" {
			C.HostDN = "cn=" + C.Override + "," + C.Base
		} else {
			C.HostDN = "cn=" + host + "," + C.Base
		}
	} else {
		C.HostDN = C.Base
	}
	Logger(nil, "Looking for files in: "+C.HostDN, "DEBUG")

	return C, nil
}
