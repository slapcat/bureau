package main

import (
	"log"
	"time"
	"text/template"
)

var (
	Tpl		*template.Template
	Files		map[string]File
)

type File struct {
	DN		string   `ldap:"dn"`
	Path		string   `ldap:"path"`
	Description	string   `ldap:"description"`
	CN		string   `ldap:"cn"`
	ObjectClass	[]string `ldap:"objectClass"`
	Data		string   `ldap:"data"`
	Perm		string   `ldap:"permissions"`
	Mtime		string	 `ldap:"modifyTimestamp"`
}

func Logger(err error, msg string, level string) {
	if err != nil && level == "FATAL" {
		log.Fatalf("[%s] %s: %v", level, msg, err)
	} else if err == nil && level == "DEBUG" {
		log.Printf("[%s] %s", level, msg)
	}
}

func ConvertLDAPtoRFC3339(mtime string) (time.Time, error) {
	const ldapLayout = "20060102150405Z"
	t, err := time.Parse(ldapLayout, mtime)

	return t, err
}
