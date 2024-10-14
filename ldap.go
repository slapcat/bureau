package main

import (
	"crypto/tls"
	"fmt"
	"github.com/go-ldap/ldap/v3"
)

func LDAPConnect() (*ldap.Conn, error) {

	l, err := ldap.DialURL(c.Server)
	if err != nil {
		return nil, err
	}

	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	return l, nil
}

func LDAPSearch(l *ldap.Conn, base string, attr []string) (*ldap.SearchResult, error) {

	l.Bind(c.Binddn, c.Password)

	searchReq := ldap.NewSearchRequest(
		base,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=configFile)",
		attr,
		nil,
	)
	result, err := l.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("Search Error: %s", err)
	}

	if len(result.Entries) > 0 {
		return result, nil
	} else {
		return nil, fmt.Errorf("Couldn't fetch search entries")
	}
}

// LDAPWrite
