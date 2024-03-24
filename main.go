package main

import (
	"reflect"
	"os"
	"log"
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
}

type File struct {
	DN string `ldap:"dn"`
	Path string `ldap:"path"`
	Description string `ldap:"description"`
	CN string `ldap:"cn"`
	ObjectClass []string `ldap:"objectClass"`
	Data string `ldap:"data"`
	Perm int `ldap:"permissions"`
}

type Kalived struct {


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


	for {
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


		for _, entry := range result.Entries {
			
			f := File{}

			err = entry.Unmarshal(&f);
			if err != nil {
				log.Fatalf("Unmarshal error: %v\n", err)
			}
			
			if slices.Contains(f.ObjectClass, "keepalivedGlobalConfig") || slices.Contains(f.ObjectClass, "keepalivedVRRPGroupConfig") || slices.Contains(f.ObjectClass, "keepalivedVRRPInstanceConfig") {
				if c.Debug { log.Printf("Generating keepalived config for %s at %s\n", f.CN, f.Path) }
				err = GenerateKeepalived(f)
				entry.PrettyPrint(1)
				log.Println(entry.GetAttributeValue("entryUUID"))
				if err != nil {
					log.Fatalf("File generation error: %v\n", err)
				}
			} else {
				if c.Debug { log.Printf("Generating default config for %s at %s\n", f.CN, f.Path) }
				err = GenerateDefault(f.Path, f.Data)
				if err != nil {
					log.Fatalf("File generation error: %v\n", err)
				}
			}
		}


		if c.Daemon {
			time.Sleep(time.Duration(c.Update) * time.Second)
		} else {
			os.Exit(0)
		}
	}
}



