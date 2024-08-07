package main

import (
	"log"
	"os"
	"time"
	"os/exec"
	"reflect"
	"gopkg.in/yaml.v3"
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

type File struct {
	DN          string   `ldap:"dn"`
	Path        string   `ldap:"path"`
	Description string   `ldap:"description"`
	CN          string   `ldap:"cn"`
	ObjectClass []string `ldap:"objectClass"`
	Data        string   `ldap:"data"`
	Perm        string   `ldap:"permissions"`
	ServiceName string   `ldap:"serviceName"`
}

var (
	paths           map[string]string
	needsUpdate     []string
	found           bool
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

		// Bind and search for timestamps
		result, err := LDAPSearch(l, c.Binddn, c.Password, hostdn, []string{"modifyTimestamp", "path"})
		if err != nil {
			log.Fatalf("LDAP search error: %v\n", err)
			continue
		}

		
		// mark files that need updating
		for _, entry := range result.Entries {
		
			mostRecent, err := compareModificationTimes(entry.GetAttributeValue("path"), entry.GetAttributeValue("modifyTimestamp"))
			if err != nil {
			  log.Printf("Could not compare modification times: %v\n", err)
			}
			
			switch mostRecent {
			case "equal":
				// nothing to do
			case "file":
				log.Printf("%s is outdated\n", entry.GetAttributeValue("path"))
				needsUpdate = append(needsUpdate, entry.DN)
				paths[entry.DN] = entry.GetAttributeValue("modifyTimestamp")
			case "ldap":
				log.Printf("%s is outdated\n", entry.DN)

				fileData, err := os.ReadFile(entry.GetAttributeValue("path"))
				if err != nil {
					log.Printf("Could not read %s: %v\n", entry.GetAttributeValue("path"), err)
				}

				err = LDAPReplace(l, entry.DN, fileData)
				if err != nil {
					log.Printf("Could not update LDAP: %v\n", err)
				}
				log.Printf("%s has been updated", entry.DN)
			}

		}

		// grab file data from LDAP
		for _, dn := range needsUpdate {

			// search for files needing updates
			result, err = LDAPSearch(l, c.Binddn, c.Password, dn, []string{})
			if err != nil {
				// In case dn changes before we can search again
				// Print error but do not exit
				log.Printf("LDAP search error: %v\n", err)
				continue
			}

			// generate files based on objectClass
			entry := result.Entries[0]
			found = false
			for _, oc := range entry.GetAttributeValues("objectClass") {

				switch oc {
				case "keepalivedGlobalConfig":

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, entry.GetAttributeValues("path"))
					}

					err = FormatKeepalived(entry, "global")
					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}

					found = true
					break
				case "keepalivedVRRPGroupConfig":

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, entry.GetAttributeValues("path"))
					}

					err = FormatKeepalived(entry, "group")
					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}

					found = true
					break
				case "keepalivedVRRPInstanceConfig":

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, entry.GetAttributeValues("path"))
					}

					err = FormatKeepalived(entry, "instance")
					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}

					found = true
					break
				}

			}

			if found != true {

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

				err = checkRestart(c.Restart, f.ServiceName)
				if err != nil {
					log.Printf("Service \"%s\" failed to restart: %v\n", f.ServiceName, err)
				}

			}

		}

		// write keepalived files
		for file, data := range KeepalivedFiles {

			if c.Debug {
				log.Printf("Writing config file to %s\n", file)
			}

			err = writeFile(file, data, "")
			if err != nil {
				log.Fatalf("File generation error: %v\n", err)
				// make this non-fatal later and then:
				// continue
			}

			err = checkRestart(c.Restart, "keepalived")
			if err != nil {
				log.Printf("Service \"%s\" failed to restart: %v\n", "keepalived", err)
			}

			delete(KeepalivedFiles, file)
		}

		l.Close()
		
		// loop if in daemon mode
		if c.Daemon {
			needsUpdate = nil
			time.Sleep(time.Duration(c.Update) * time.Second)
		} else {
			os.Exit(0)
		}

	}
}

func compareModificationTimes(path string, timestamp string) (string, error) {

	// parse file mod time
	fileInfo, err := os.Stat(path) 
	if err != nil { 
		log.Fatalf("File read error for %s: %v\n", path, err) 
  } 
	fileMTime := fileInfo.ModTime().Truncate(60 * time.Second)

	// parse ldap mod time
	ldapMTime, _ := time.Parse("20060102150405Z0700", timestamp)
	ldapMTime = ldapMTime.Truncate(60 * time.Second)

	// compare
	if fileMTime.Equal(ldapMTime) {
		return "equal", nil
	} else if ldapMTime.After(fileMTime) {
		return "file", nil
	} else {
		return "ldap", nil
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
	cmd := exec.Command("chmod", perm, path+".tmp")
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
	err = os.Rename(path+".tmp", path)
	if err != nil {
		return err
	}

	return nil
}

func findConfig() ([]byte, error) {

	var conf string

	home, err := os.UserHomeDir()
	locations := []string{"bureau.yaml",
		home + "/.bureau.yaml",
		home + "/.config/bureau/bureau.yaml",
		"/etc/bureau/bureau.yaml"}

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

func checkRestart(enabled bool, svc string) error {

	if enabled && svc != "" {
		cmd := exec.Command("systemctl", "restart", svc)
		cmd.Stderr = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil

}
