package main

import (
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"reflect"
	"time"
//	"bytes"
)

type Config struct {
	Debug    bool `default:"false"`
	Daemon   bool `default:"true"`
	Server   string
	Binddn   string
	Password string
	Base     string
	Update   int    `yaml:"update_interval" default:"600"`
	Host     bool   `yaml:"host_specific_entries" default:"true"`
	Restart  bool   `yaml:"restart_service_on_change" default:"true"`
	Override string `yaml:"override_hostname"`
}

var (
	paths map[string]string
	needsUpdate []string
	KeepalivedFiles map[string]string
)

func main() {

	// load bureau config
	conf, err := findConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	c := Config{}
	err = yaml.Unmarshal([]byte(conf), &c)
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
	if c.Debug {
		log.Printf(" === Looking for files in: %s", hostdn)
	}

	// init paths map for tracking updates and data
	paths = make(map[string]string)
  KeepalivedFiles = make(map[string]string)

	for {

		// LDAP connect
		l, err := LDAPConnect(c.Server)
		if err != nil {
			log.Fatalf("Connection error: %v\n", err)
		}
		defer l.Close()

		// Bind and search for timestamps
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
				// no need for the for loop because we pull one at a time like this:
				log.Println(result.Entries[0])
				//				if slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedGlobalConfig") || slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedVRRPInstanceConfig") || slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedVRRPGroupConfig") {
				if slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedGlobalConfig") {

					f := Kalived{}
					err = entry.Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, f.Path)
					}

					err = FormatKeepalivedGlobal(f)
					// debugging
					//log.Println(KeepalivedGlobal.String())

					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}

					if c.Debug {
            log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, f.Path)
          }
				
				} else {

					f := File{}
					err = entry.Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Writing config file for %s at %s\n", f.CN, f.Path)
					}

					err = writeFile(f.Path, f.Data, f.Perm)
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

func writeFile(path string, data string, perm string) error {

	// create tmp file
	f, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}

	// set permissions
	if perm == "" {
		perm = "0600"
	}

	// use exec because of type issue with os.Chmod
	cmd := exec.Command("chmod", perm, path + ".tmp")
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	// write to tmp file
	_, err = f.WriteString(data)
	defer f.Close()
	if err != nil {
		return err
	}

	// move into place if write is good
	err = os.Rename(path + ".tmp", path)
	if err != nil {
		return err
	}

	return nil
}

func findConfig() ([]byte, error) {

	var conf string

	home, err := os.UserHomeDir()
	locations := []string{home + "/.bureau.yaml",
		home + "/.config/bureau/bureau.yaml",
		"/etc/bureau/bureau.yaml", "bureau.yaml"}

	for _, path := range locations {
		if _, err = os.Stat(path); err == nil {
			conf = path
			break
		}
	}

	data, err := os.ReadFile(conf)
	if err != nil {
		return nil, err
	}

	return data, nil
}
