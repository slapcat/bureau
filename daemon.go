package main

import (
	"regexp"
	"strings"
)

func Summon() {

	// init Files map
	Files = make(map[string]File)

	// Define regex
	hasKeepalivedOC := regexp.MustCompile("keepalived")

	// LDAP connect
	l, err := LDAPConnect()
	Logger(err, "Connection error", "FATAL")

	defer l.Close()

	// Bind and search for files that need updating
	result, err := LDAPSearch(l, c.HostDN, []string{"modifyTimestamp", "path"})
	Logger(err, "LDAP search error", "WARN")

	// mark files that need updating
	for _, entry := range result.Entries {

		// Import to struct
		f := File{}
		err = entry.Unmarshal(&f)
		Logger(err, "Unmarshal error", "FATAL")

		// Check for file and get mtime
		exists, fileOnDiskMtime := FileExist(f.Path)
		if ! exists {
			Files[f.DN] = f
			continue
		}

		// Get ldap mtime
		ldapMtime, err := ConvertLDAPtoRFC3339(f.Mtime)
		if err != nil {
			Logger(err, "Failed converting LDAP time to RFC3339", "WARN")
			continue
		}   

		// Compare LDAP and file mtime
		if ldapMtime.After(fileOnDiskMtime) {
			Logger(nil, "Queueing outdated file: "+f.DN, "DEBUG")
			Files[f.DN] = f
		}
	}


	// Grab file data from LDAP for files needing update
	for _, file := range Files {

		// Request remaining attributes from LDAP and add to struct
		result, err = LDAPSearch(l, file.DN, []string{"*", "+"})
		if err != nil {
			Logger(err, "LDAP search error", "WARN")
			continue
		}

		err = result.Entries[0].Unmarshal(&file)
		if err != nil {
			Logger(err, "Unmarshal error", "FATAL")
		}


		// Convert objectClass slice to byte slice
		objectClassesString:= strings.Join(file.ObjectClass, "")
		objectClasses := []byte(objectClassesString)
			
		// Format data based on objectClass
		switch {
			case hasKeepalivedOC.Match(objectClasses):
				f := Kalived{}
				err = result.Entries[0].Unmarshal(&f)
				Logger(err, "Unmarshal error", "FATAL")

				Logger(nil, "Writing keepalived instance ("+f.InstanceName+"): "+f.Path, "DEBUG")

				err = FormatKeepalived(f, "kinstance")
				Logger(err, "Formatting error", "WARN")

				break

			default:
				f := Files[file.DN]
				err = result.Entries[0].Unmarshal(&f)
				Logger(err, "Unmarshal error", "FATAL")

				Logger(nil, "Writing generic config file to "+f.Path, "DEBUG")

				mtime, err := ConvertLDAPtoRFC3339(f.Mtime)
				if err != nil {
					Logger(err, "Failed converting LDAP time to RFC3339", "WARN")
					continue
				}

				err = WriteFile(f.Path, f.Data, f.Perm, mtime)
				Logger(err, "File generation error", "FATAL")
		}
	}

	PublishFiles()
}
