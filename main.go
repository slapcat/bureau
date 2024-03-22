package main

import (
	"reflect"
	"os"
	"log"
	"strconv"
	"golang.org/x/exp/slices"
	"time"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug bool
	Daemon bool
	Server string
	Binddn string
	Password string
	Base string
	Update int `yaml:"update_interval"`
	Host bool `yaml:"host_specific_entries"`
	Restart bool `yaml:"restart_service_on_change"`
	Override string `yaml:"override_hostname"`
	Files []string
}

type File struct {
	DN string `ldap:"dn"`
	Path string `ldap:"path"`
	Description string `ldap:"description"`
	CN string `ldap:"cn"`
	ObjectClass []string `ldap:"objectClass"`
	Data string `ldap:"data"`
	Perm string `ldap:"permissions"`
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
	} else if c.Debug {
		for i := 0; i <= 10; i++ {
			key := reflect.Indirect(reflect.ValueOf(c)).Type().Field(i).Name
			value := reflect.ValueOf(c)
			log.Printf(" === Loading configuration %v: %v\n", key, value.FieldByName(key))
		}
	}

	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("error: %v", err)
	} else if c.Debug {
		log.Printf(" === Getting hostname: %s", host)
	}

	var hostdn string
	if c.Host {
		if c.Override != "" {
			hostdn = "cn=" + c.Override + "," + c.Base
		} else {
			hostdn = "cn=" + host + "," + c.Base
		}
	} else {
		hostdn = c.Base
	}
	if c.Debug { log.Printf(" === Looking for files in: %s", hostdn) }

	// Non-TLS Connection
	l, err := LDAPConnect(c.Server)
	if err != nil {
		log.Fatalf("Connection error: %v\n", err)
	}
	defer l.Close()

	result, err := LDAPSearch(l, c.Binddn, c.Password, hostdn)
	if err != nil {
		log.Fatalf("LDAP search error: %v\n", err)
	}

	f := File{}

	for _, entry := range result.Entries {

		err = entry.Unmarshal(&f);
		if err != nil {
			log.Fatalf("Unmarshal error: %v\n", err)
		}

		// NOT WORKING
		var perm uint64
		if f.Perm == "" {
			perm = 0600
		} else {
			perm, err = strconv.ParseUint(f.Perm, 0, 32) 
		}
		
		if slices.Contains(f.ObjectClass, "keepalivedConfig") {
			if c.Debug { log.Printf("Generating keepalived config for %s at %s\n", f.CN, f.Path) }
			// _, err = GenerateKeepalived(f)
		} else {
			if c.Debug { log.Printf("Generating default config for %s at %s\n", f.CN, f.Path) }
			err = GenerateDefault(f.Path, f.Data, perm)
			if err != nil {
				log.Fatalf("File generation error: %v\n", err)
			}
		}
	}

	time.Sleep(time.Duration(c.Update) * time.Second)
}



