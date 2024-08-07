package main

import (
	"fmt"
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
)

func LDAPConnect(host string) (*ldap.Conn, error) {

	l, err := ldap.DialURL(host)
	if err != nil {
		return nil, err
	}

	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	return l, nil
}

func LDAPSearch(l *ldap.Conn, binddn string, password string, base string, attr []string) (*ldap.SearchResult, error) {

	l.Bind(binddn, password)

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

func LDAPReplace(l *ldap.Conn, DN string, data []byte) error {

	modify := ldap.NewModifyRequest(DN, nil)

	modify.Replace("data", []string{string(data)})
	err := l.Modify(modify)
	if err != nil {
			return err
	}

	return nil
}
