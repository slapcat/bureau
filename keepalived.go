package main

import (
	"bytes"
	"text/template"
)

type Kalived struct {
	DN			    	string   `ldap:"dn"`
	Path                        	string   `ldap:"path"`
	Perm                        	string   `ldap:"permissions"`
	Mtime                       	string   `ldap:"modifyTimestamp"`
	ServiceName                 	string   `ldap:"serviceName"`
	GlobalNotificationEmail     	[]string `ldap:"globalNotificationEmail"`
	GlobalNotificationEmailFrom 	string   `ldap:"globalNotificationEmailFrom"`
	GlobalSMTPServer            	string   `ldap:"globalSMTPServer"`
	GlobalSMTPConnectTimeout    	int      `ldap:"globalSMTPConnectTimeout"`
	GlobalLVSId                 	string   `ldap:"globalLVSId"`
	GroupName                   	string   `ldap:"groupName"`
	GroupMember                 	[]string `ldap:"groupMember"`
	NotifyMasterVRRPGroup       	string   `ldap:"notifyMasterVRRPGroup"`
	NotifyBackupVRRPGroup       	string   `ldap:"notifyBackupVRRPGroup"`
	NotifyFaultVRRPGroup        	string   `ldap:"notifyFaultVRRPGroup"`
	InstanceName                	string   `ldap:"instanceName"`
	/* Need to treat next value as string since
	entry.Unmarshal method doesn't support bool */
	SMTPAlert                	string   `ldap:"smtpAlert"`
	AuthType                 	string   `ldap:"authType"`
	AuthPass                 	string   `ldap:"authPass"`
	VirtualIPAddress         	[]string `ldap:"virtualIPAddress"`
	VirtualIPAddressExcluded 	[]string `ldap:"virtualIPAddressExcluded"`
	State                    	string   `ldap:"state"`
	Interface                	string   `ldap:"interface"`
	McastSrcIP               	string   `ldap:"mcastSrcIP"`
	LVSSyncDaemonInterface   	string   `ldap:"lvsSyncDaemonInterface"`
	VirtualRouterID          	int      `ldap:"virtualRouterID"`
	Priority                 	int      `ldap:"priority"`
	AdvertInt                	int      `ldap:"advertInt"`
	UnicastSrcIP             	string   `ldap:"unicastSrcIP"`
	UnicastPeer              	[]string `ldap:"unicastPeer"`
	NotifyMasterVRRPInstance 	string   `ldap:"notifyMasterVRRPInstance"`
	NotifyBackupVRRPInstance 	string   `ldap:"notifyBackupVRRPInstance"`
	NotifyFaultVRRPInstance  	string   `ldap:"notifyFaultVRRPInstance"`
}

func FormatKeepalived(inter any, class string) error {

	f := inter.(Kalived)

	if Tpl == nil || Tpl.Lookup(class+".tmpl") == nil {
		Tpl = template.Must(template.ParseGlob("templates/keepalived/*.tmpl"))
	}

	var configData bytes.Buffer
	err := Tpl.ExecuteTemplate(&configData, class+".tmpl", f)
	Logger(err, "Template creation error", "FATAL")

	mtime, err := ConvertLDAPtoRFC3339(f.Mtime)
	Logger(err, "Failed converting LDAP time to RFC3339", "WARN")

	WriteFile(f.Path, configData.String(), f.Perm, mtime)

	return nil
}
