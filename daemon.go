package main

import (
	"log"
)

var (
	paths						map[string]string
	needsUpdate			[]string
	found						bool
	KeepalivedFiles map[string]string
)

func Summon() {

	// init paths map for tracking updates and data
	paths = make(map[string]string)
  KeepalivedFiles = make(map[string]string)

	// LDAP connect
	l, err := LDAPConnect()
	if err != nil {
		log.Fatalf("Connection error: %v\n", err)
	}
	defer l.Close()

	// Bind and search for timestamps
	result, err := LDAPSearch(l, c.HostDN, []string{"modifyTimestamp"})
	if err != nil {
		log.Fatalf("LDAP search error: %v\n", err)
		//continue
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

		// search for files needing updates
		result, err = LDAPSearch(l, dn, []string{})
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
					f := Kalived{}
					err = entry.Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, f.Path)
					}

					err = FormatKeepalived(f, "global")

					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}

					found = true
					break
				case "keepalivedVRRPGroupConfig":
					f := Kalived{}
					err = entry.Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", entry.DN, f.Path)
					}

					err = FormatKeepalived(f, "group")

					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}					

					found = true
					break
				case "keepalivedVRRPInstanceConfig":
					//format
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

				err = WriteFile(f.Path, f.Data, f.Perm)
				if err != nil {
					log.Fatalf("File generation error: %v\n", err)
				}

		}

	}


	// write keepalived files
	for file, data := range KeepalivedFiles {

		if c.Debug {
			log.Printf("Writing config file to %s\n", file)
		}

		WriteFile(file, data, "")

	}

	needsUpdate = nil
}
