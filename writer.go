package main

import (
	"os"
	"time"
)

func FileExist(path string) (bool, time.Time) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			Logger(err, "File does not exist", "DEBUG")
			return false, time.Time{}
		}
	}

	return true, info.ModTime()
}

func WriteFile(path string, data string, perm string, mtime time.Time) error {
	// set permissions
	if perm == "" {
		perm = "0600"
	}

	// create and open tmp file
	f, err := os.OpenFile(path + ".tmp", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	Logger(err, "Error opening file", "FATAL")

	// write to tmp file
	_, err = f.WriteString(data)
	Logger(err, "Failed to write to tmp file", "FATAL")

	defer f.Close()

	return nil
}

func PublishFiles() error {
	movedFiles := make(map[string]bool)

	for _, f := range Files {

		if _, ok := movedFiles[f.Path]; ok {
			// already moved this path
			continue
		}

		// Move tmp file
		err := os.Rename(f.Path+".tmp", f.Path)
		Logger(err, "Error publishing file", "WARN")
		movedFiles[f.Path] = true

		// Set file timestamps to now
		mtime := time.Now()
		atime := mtime
		//mtime, err := ConvertLDAPtoRFC3339(f.Mtime)
		//Logger(err, "Failed converting LDAP time to RFC3339", "WARN")

		err = os.Chtimes(f.Path, atime, mtime)
		Logger(err, "Error updating file properties", "WARN")

		delete(Files, f.DN)
	}

	return nil
}
