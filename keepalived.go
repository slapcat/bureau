package main

import (
	"fmt"
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
	ServiceName                 string   `ldap:"serviceName"`
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
	UnicastSrcIP             string   `ldap:"unicastSrcIP"`
	UnicastPeer              []string `ldap:"unicastPeer"`
	NotifyMasterVRRPInstance string   `ldap:"notifyMasterVRRPInstance"`
	NotifyBackupVRRPInstance string   `ldap:"notifyBackupVRRPInstance"`
	NotifyFaultVRRPInstance  string   `ldap:"notifyFaultVRRPInstance"`
}

func FormatKeepalived(inter any, class string) error {

	f := inter.(Kalived)
	var tmpl *template.Template
	var err error

	switch class {
	case "global":
		tmpl, err = template.New("kglobal").Parse(`
global_defs {
	{{if .GlobalNotificationEmail}}notification_email {
		{{range .GlobalNotificationEmail}}{{.}}
		{{end}}
	}{{end}}
	{{if .GlobalNotificationEmail}}notification_email_from {{.GlobalNotificationEmailFrom}}{{end}}
	{{if .GlobalSMTPServer}}smtp_server {{.GlobalSMTPServer}}{{end}}
	{{if .GlobalSMTPConnectTimeout}}smtp_connect_timeout {{.GlobalSMTPConnectTimeout}}{{end}}
	{{if .GlobalLVSId}}lvs_id {{.GlobalLVSId}}{{end}}
}
`)

	case "group":
		tmpl, err = template.New("kgroup").Parse(`
vrrp_sync_group {{.GroupName}} {
	group {
		{{range .GroupMember}}{{.}}
		{{end}}
	}
	{{if .NotifyMasterVRRPGroup}}notify_master {{.NotifyMasterVRRPGroup}}
	{{end}}{{if .NotifyBackupVRRPGroup}}notify_backup {{.NotifyBackupVRRPGroup}}
	{{end}}{{if .NotifyFaultVRRPGroup}}notify_fault {{.NotifyFaultVRRPGroup}}{{end}}
}
`)

	case "instance":
		tmpl, err = template.New("kinstance").Parse(`
vrrp_instance {{.InstanceName}} {
	state {{.State}}
	interface {{.Interface}}
	{{if .McastSrcIP}}mcast_src_ip {{.McastSrcIP}}{{end}}
	{{if .UnicastSrcIP}}unicast_src_ip {{.UnicastSrcIP}}
	unicast_peer {
		{{range .UnicastPeer}}{{.}}
		{{end -}}
	}{{end}}
	{{if .LVSSyncDaemonInterface}}lvs_sync_daemon_interface {{.LVSSyncDaemonInterface}}{{end}}
	virtual_router_id {{.VirtualRouterID}}
	priority {{.Priority}}
	{{if .AdvertInt}}advert_int {{.AdvertInt}}{{end}}
	{{if .SMTPAlert}}smtp_alert {{.SMTPAlert}}{{end}}
	authentication {
		auth_type {{.AuthType}}
		auth_pass {{.AuthPass}}
	}
	virtual_ipaddress {
		{{range .VirtualIPAddress}}{{.}}
	{{end -}}
	}
	{{if .VirtualIPAddressExcluded}}}virtual_ipaddress_excluded {
	  {{range .VirtualIPAddressExcluded}}{{.}}
		{{end -}}
	}{{end}}
	{{if .NotifyMasterVRRPInstance}}notify_master {{.NotifyMasterVRRPInstance}}
	{{end}}{{if .NotifyBackupVRRPInstance}}notify_backup {{.NotifyBackupVRRPInstance}}
	{{end}}{{if .NotifyFaultVRRPInstance}}notify_fault {{.NotifyFaultVRRPInstance}}{{end}}
}
`)

	default:
		return fmt.Errorf("Template not found: %s\n", class)
	}

	if err != nil {
		log.Fatalf("Template creation error: %v\n", err)
	}

	var newData bytes.Buffer
	err = tmpl.Execute(&newData, f)
	if err != nil {
		log.Fatalf("Template creation error: %v\n", err)
	}

	KeepalivedFiles[f.Path] = KeepalivedFiles[f.Path] + newData.String()

	return nil
}
