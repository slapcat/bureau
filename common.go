package main

type File struct {
	DN          string   `ldap:"dn"`
	Path        string   `ldap:"path"`
	Description string   `ldap:"description"`
	CN          string   `ldap:"cn"`
	ObjectClass []string `ldap:"objectClass"`
	Data        string   `ldap:"data"`
	Perm        string   `ldap:"permissions"`
	Mtime				string	 `ldap:"modifyTimestamp"`
}
