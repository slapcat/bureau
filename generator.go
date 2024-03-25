package main

import (
	"os"
	"os/exec"
	"log"
)

func GenerateDefault(path string, data string, perm string) error {

	// create file
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	// set permissions
	if perm == "" {
		perm = "0600"
	}

	cmd := exec.Command( "chmod", perm, path )
  cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	// write file
	_, err = f.WriteString(data)
	defer f.Close()
	if err != nil {
		return err
	}

	return nil
}

func GenerateKeepalived(inter interface{}) error {


	f := inter.(Kalived)
	log.Println(f)
	return nil
	
}
