package main

import (
	"os"
	"os/exec"
)

func WriteFile(path string, data string, perm string) error {

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
