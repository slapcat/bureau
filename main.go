package main

import (
	"os"
	"log"
	"time"
	"reflect"
	"gopkg.in/yaml.v3"
	"golang.org/x/exp/slices"
)

type Config struct {
	Debug bool `default:"false"`
	Daemon bool `default:"true"`
	Server string
	Binddn string
	Password string
	Base string
	Update int `yaml:"update_interval" default:"600"`
	Host bool `yaml:"host_specific_entries" default:"true"`
	Restart bool `yaml:"restart_service_on_change" default:"true"`
	Override string `yaml:"override_hostname"`
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

type Kalived struct {


}

var paths map[string]string
var needsUpdate []string

func main() {

	// load bureau config
	data, err := os.ReadFile("bureau.yaml")
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

	// get hostname and set search base
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

	// init paths map for tracking updates
	paths = make(map[string]string)

	for {

		// Non-TLS Connection
		l, err := LDAPConnect(c.Server)
		if err != nil {
			log.Fatalf("Connection error: %v\n", err)
		}
		defer l.Close()

		// search for timestamps
		result, err := LDAPSearch(l, c.Binddn, c.Password, hostdn, []string{"modifyTimestamp"})
		if err != nil {
			log.Fatalf("LDAP search error: %v\n", err)
			continue
		}

		// mark files that need updating
		for _, entry := range result.Entries {
			if val, ok := paths[entry.DN]; ok {
				if entry.GetAttributeValue("modifyTimestamp") == val {
					// no update needed
					continue
				}
			}

			log.Printf("%s is outdated\n", entry.DN)
			needsUpdate = append(needsUpdate, entry.DN)
			paths[entry.DN] = entry.GetAttributeValue("modifyTimestamp")
			
		}
	
		// grab file data from LDAP
		for _, dn := range needsUpdate {

			result, err = LDAPSearch(l, c.Binddn, c.Password, dn, []string{})
			if err != nil {
				// In case dn changes before we can search again
				// Print error but do not exit
				log.Printf("LDAP search error: %v\n", err)
				continue
			}

			// generate files based on objectClass
			for _, entry := range result.Entries {
				
				f := File{}

				err = entry.Unmarshal(&f);
				if err != nil {
					log.Fatalf("Unmarshal error: %v\n", err)
				}
				
				if slices.Contains(f.ObjectClass, "keepalivedGlobalConfig") ||
				slices.Contains(f.ObjectClass, "keepalivedVRRPGroupConfig") ||
				slices.Contains(f.ObjectClass, "keepalivedVRRPInstanceConfig") {
					if c.Debug { log.Printf("Generating keepalived config for %s at %s\n", f.CN, f.Path) }
					err = GenerateKeepalived(f)
					if err != nil {
						log.Fatalf("File generation error: %v\n", err)
					}
				} else {
					if c.Debug { log.Printf("Generating config file for %s at %s\n", f.CN, f.Path) }
					err = GenerateDefault(f.Path, f.Data, f.Perm)
					if err != nil {
						log.Fatalf("File generation error: %v\n", err)
					}
				}
			}
		}

		// loop if in daemon mode
		if c.Daemon {
			needsUpdate = nil
			time.Sleep(time.Duration(c.Update) * time.Second)
		} else {
			os.Exit(0)
		}

	}
}


