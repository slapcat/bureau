package main

import (
	"os"
	"log"
	"time"
)

var (
	Files						map[string]File
)

func Summon() {

	// init paths map for tracking updates and data
	Files = make(map[string]File)
  KeepalivedFiles = make(map[string]string)

	// LDAP connect
	l, err := LDAPConnect()
	if err != nil {
		log.Fatalf("Connection error: %v\n", err)
	}
	defer l.Close()

	// Bind and search for files that need updating
	result, err := LDAPSearch(l, c.HostDN, []string{"modifyTimestamp", "path"})
	if err != nil {
		log.Printf("LDAP search error: %v\n", err)
	}

	// mark files that need updating
	for _, entry := range result.Entries {

		// Import to struct
		f := File{}
		err = entry.Unmarshal(&f)
		if err != nil {
			log.Fatalf("Unmarshal error: %v\n", err)
		}

		// Get file mtime
		info, err := os.Stat(f.Path)
   	if err != nil {
			log.Printf("Could not get file info: %v\n", err)

			// File doesn't exist, schedule update
			if os.IsNotExist(err) {
				Files[f.DN] = f
				break
			}
		}

		// Get ldap mtime
		mtime, err := convertLDAPtoRFC3339(f.Mtime)
		if err != nil {
			log.Printf("Failed converting LDAP time to RFC3339: %v\n", err)
			continue
		}   

		// Compare LDAP and file mtime
		if mtime.After(info.ModTime()) {
			log.Printf("%s is outdated\n", f.DN)
			Files[f.DN] = f
		}
	}


	// Grab file data from LDAP for files needing update
	for _, file := range Files {

		// Request remaining attributes from LDAP and add to struct
		result, err = LDAPSearch(l, file.DN, []string{"*", "+"})
		if err != nil {
			log.Printf("LDAP search error: %v\n", err)
			continue
		}

		err = result.Entries[0].Unmarshal(&file)
		if err != nil {
			log.Fatalf("Unmarshal error: %v\n", err)
		}

		// generate files based on objectClass
		for i, oc := range file.ObjectClass {
			
			switch oc {

				case "keepalivedGlobalConfig":
					f := Kalived{}
					err = result.Entries[0].Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", f.DN, f.Path)
					}

					err = FormatKeepalived(f, "global")
					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}

					break
				case "keepalivedVRRPGroupConfig":
					f := Kalived{}
					err = result.Entries[0].Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", f.DN, f.Path)
					}

					err = FormatKeepalived(f, "group")

					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}					

					break
				case "keepalivedVRRPInstanceConfig":
					f := Kalived{}
					err = result.Entries[0].Unmarshal(&f)
					if err != nil {
						log.Fatalf("Unmarshal error: %v\n", err)
					}

					if c.Debug {
						log.Printf("Formatting keepalived config for %s at %s\n", f.DN, f.Path)
					}

					err = FormatKeepalived(f, "instance")

					if err != nil {
						log.Fatalf("Formatting error: %v\n", err)
					}					

					break

				default:

					if i == len(file.ObjectClass) - 1 {
						f := Files[file.DN]
						err = result.Entries[0].Unmarshal(&f)
						if err != nil {
							log.Fatalf("Unmarshal error: %v\n", err)
						}

						if c.Debug {
							log.Printf("Writing config file for %s at %s\n", f.CN, f.Path)
						}   

						mtime, err := convertLDAPtoRFC3339(f.Mtime)
						if err != nil {
							log.Printf("Failed converting LDAP time to RFC3339: %v\n", err)
							continue
						}   

						err = WriteFile(f.Path, f.Data, f.Perm, mtime)
						if err != nil {
							log.Fatalf("File generation error: %v\n", err)
						}   
					
						delete(Files, f.DN)
					}
			}

		}
	}


	for _, kfile := range KeepalivedMap {

		if c.Debug {
			log.Printf("Writing config file to %s\n", kfile.Path)
		}

		mtime, err := convertLDAPtoRFC3339(kfile.Mtime)
		if err != nil {
			log.Printf("Failed converting LDAP time to RFC3339: %v\n", err)
			continue
		}   

		WriteFile(kfile.Path, KeepalivedFiles[kfile.Path], "", mtime)

		// remove all LDAP entries related to this file path
		for dn, data := range KeepalivedMap {
			if data.Path == kfile.Path {
				delete(KeepalivedMap, dn)
			}
		}
	}
	
}

func convertLDAPtoRFC3339(mtime string) (time.Time, error) {
	const ldapLayout = "20060102150405Z"
	t, err := time.Parse(ldapLayout, mtime)

	return t, err
}
