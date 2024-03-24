package main

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
)

func LDAPConnect(host string) (*ldap.Conn, error) {

	l, err := ldap.DialURL(host)
	if err != nil {
		return nil, err
	}

	return l, nil
}

/* func LDAPCheck() {

    searchReq := ldap.NewSearchRequest(
        base,
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0,
        0,
        false,
        "(objectClass=configFile)",
        []string{"modifyTimestamp"},
        nil,
    )

compare timestamp with existing one if any to determine which entries to pull


}

*/


func LDAPSearch(l *ldap.Conn, binddn string, password string, base string) (*ldap.SearchResult, error) {
	l.Bind(binddn, password)

    searchReq := ldap.NewSearchRequest(
        base,
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0,
        0,
        false,
        "(objectClass=configFile)",
        []string{"modifyTimestamp", "*"},
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
