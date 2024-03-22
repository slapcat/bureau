package main

import "os"
import "log"

func GenerateDefault(path string, data string, perm uint64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	err = os.Chmod(path, os.FileMode(perm))
	if err != nil {
		return err
	}

	_, err = f.WriteString(data)
	defer f.Close()
	if err != nil {
		return err
	}

	return nil
}

func GenerateKeepalived(file interface{}) error {

	f := file.(File)
	log.Println(f)
	return nil
}
