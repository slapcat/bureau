package main

import (
	"log"
	"bytes"
	"text/template"
)

type File struct {
	DN          string   `ldap:"dn"`
	Path        string   `ldap:"path"`
	Description string   `ldap:"description"`
	CN          string   `ldap:"cn"`
	ObjectClass []string `ldap:"objectClass"`
	Data        string   `ldap:"data"`
	Perm        string   `ldap:"permissions"`
}

type Kalived struct {
	Path                        string   `ldap:"path"`
	Perm                        string   `ldap:"permissions"`
	GlobalNotificationEmail     []string `ldap:"globalNotificationEmail"`
	GlobalNotificationEmailFrom string   `ldap:"globalNotificationEmailFrom"`
	GlobalSMTPServer            string   `ldap:"globalSMTPServer"`
	GlobalSMTPConnectTimeout    int      `ldap:"globalSMTPConnectTimeout"`
	GlobalLVSId                 string   `ldap:"globalLVSId"`
	GroupName                   string   `ldap:"groupName"`
	GroupMember                 []string `ldap:"groupMember"`
	NotifyMasterVRRPGroup       string   `ldap:"notifyMasterVRRPGroup"`
	NotifyBackupVRRPGroup       string   `ldap:"notifyBackupVRRPGroup"`
	NotifyFaultVRRPGroup        string   `ldap:"notifyFaultVRRPGroup"`
	InstanceName                string   `ldap:"instanceName"`
	/* Need to treat next value as string since
	entry.Unmarshal method doesn't support bool */
	SMTPAlert                string   `ldap:"smtpAlert"`
	AuthType                 string   `ldap:"authType"`
	AuthPass                 string   `ldap:"authPass"`
	VirtualIPAddress         []string `ldap:"virtualIPAddress"`
	VirtualIPAddressExcluded []string `ldap:"virtualIPAddressExcluded"`
	State                    string   `ldap:"state"`
	Interface                string   `ldap:"interface"`
	McastSrcIP               string   `ldap:"mcastSrcIP"`
	LVSSyncDaemonInterface   string   `ldap:"lvsSyncDaemonInterface"`
	VirtualRouterID          int      `ldap:"virtualRouterID"`
	Priority                 int      `ldap:"priority"`
	AdvertInt                int      `ldap:"advertInt"`
}

func FormatKeepalivedGlobal(inter any) error {

	f := inter.(Kalived)

	tmpl, err := template.New("kglobal").Parse(`
	global_defs {
 		  {{if .GlobalNotificationEmail}}notification_email {
			{{range .GlobalNotificationEmail}}{{.}}
			{{end}}}{{end}}
			{{if .GlobalNotificationEmail}}notification_email_from {{.GlobalNotificationEmailFrom}}{{end}}
			{{if .GlobalSMTPServer}}smtp_server {{.GlobalSMTPServer}}{{end}}
			{{if .GlobalSMTPConnectTimeout}}smtp_connect_timeout {{.GlobalSMTPConnectTimeout}}{{end}}
			{{if .GlobalLVSId}}lvs_id {{.GlobalLVSId}}{{end}}
	}
	`)
	if err != nil {
		log.Fatalf("Template creation error: %v\n", err)
	}

	var newData bytes.Buffer
	err = tmpl.Execute(&newData, inter.(Kalived))
	if err != nil {
		log.Fatalf("Template creation error: %v\n", err)
	}

	KeepalivedFiles[f.Path] = KeepalivedFiles[f.Path] + newData.String()

	return nil
}
