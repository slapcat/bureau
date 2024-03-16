package main

import "os"

func GenerateDefault(path string, data string) error {
	f, err := os.Create(path)

	if err != nil {


		return err

	}

	_, err = f.WriteString(data)

	if err != nil {


        f.Close()

		return err

	}


        f.Close()

	return nil
}
