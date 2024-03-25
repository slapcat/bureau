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
	Path string `ldap:"path"`
	Perm string `ldap:"permissions"`
	GlobalNotificationEmail	[]string `ldap:"globalNotificationEmail"`
	GlobalNotificationEmailFrom string `ldap:"globalNotificationEmailFrom"`
	GlobalSMTPServer string `ldap:"globalSMTPServer"`
	GlobalSMTPConnectTimeout int `ldap:"globalSMTPConnectTimeout"`
	GlobalLVSId string `ldap:"globalLVSId"`
	GroupName string `ldap:"groupName"`
	GroupMember []string `ldap:"groupMember"`
	NotifyMasterVRRPGroup string `ldap:"notifyMasterVRRPGroup"`
	NotifyBackupVRRPGroup string `ldap:"notifyBackupVRRPGroup"`
	NotifyFaultVRRPGroup string `ldap:"notifyFaultVRRPGroup"`
	InstanceName string `ldap:"instanceName"`
	/* Need to treat next value as string since 
	entry.Unmarshal method doesn't support bool */
	SMTPAlert string `ldap:"smtpAlert"`
	AuthType string `ldap:"authType"`
	AuthPass string `ldap:"authPass"`
	VirtualIPAddress []string `ldap:"virtualIPAddress"`
	VirtualIPAddressExcluded []string `ldap:"virtualIPAddressExcluded"`
	State string `ldap:"state"`
	Interface string `ldap:"interface"`
	McastSrcIP string `ldap:"mcastSrcIP"`
	LVSSyncDaemonInterface string `ldap:"lvsSyncDaemonInterface"`
	VirtualRouterID int `ldap:"virtualRouterID"`
	Priority int `ldap:"priority"`
	AdvertInt int `ldap:"advertInt"`
}

var paths map[string]string
var needsUpdate []string

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
	if c.Debug { log.Printf(" === Looking for files in: %s", hostdn) }

	// init paths map for tracking updates
	paths = make(map[string]string)

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
				if slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedGlobalConfig") || slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedVRRPInstanceConfig") || slices.Contains(entry.GetAttributeValues("objectClass"), "keepalivedVRRPGroupConfig") {

					f := Kalived{}
					err = entry.Unmarshal(&f);
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug { log.Printf("Generating keepalived config for %s at %s\n", entry.DN, f.Path) }
	
					err = GenerateKeepalived(f)
					if err != nil {
						log.Fatalf("File generation error: %v\n", err)
					}
				
				} else {

					f := File{}
					err = entry.Unmarshal(&f);
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}
			
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
