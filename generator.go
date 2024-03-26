package main

import (
	"os"
	"os/exec"
	"log"
)

type File struct {
	DN string `ldap:"dn"`
	Path string `ldap:"path"`
	Description string `ldap:"description"`
	CN string `ldap:"cn"`
	ObjectClass []string `ldap:"objectClass"`
	Data string `ldap:"data"`
	Perm string `ldap:"permissions"`
}

type Kalived struct {
	Path string `ldap:"path"`
	Perm string `ldap:"permissions"`
	GlobalNotificationEmail	[]string `ldap:"globalNotificationEmail"`
	GlobalNotificationEmailFrom string `ldap:"globalNotificationEmailFrom"`
	GlobalSMTPServer string `ldap:"globalSMTPServer"`
	GlobalSMTPConnectTimeout int `ldap:"globalSMTPConnectTimeout"`
	GlobalLVSId string `ldap:"globalLVSId"`
	GroupName string `ldap:"groupName"`
	GroupMember []string `ldap:"groupMember"`
	NotifyMasterVRRPGroup string `ldap:"notifyMasterVRRPGroup"`
	NotifyBackupVRRPGroup string `ldap:"notifyBackupVRRPGroup"`
	NotifyFaultVRRPGroup string `ldap:"notifyFaultVRRPGroup"`
	InstanceName string `ldap:"instanceName"`
	/* Need to treat next value as string since
	entry.Unmarshal method doesn't support bool */
	SMTPAlert string `ldap:"smtpAlert"`
	AuthType string `ldap:"authType"`
	AuthPass string `ldap:"authPass"`
	VirtualIPAddress []string `ldap:"virtualIPAddress"`
	VirtualIPAddressExcluded []string `ldap:"virtualIPAddressExcluded"`
	State string `ldap:"state"`
	Interface string `ldap:"interface"`
	McastSrcIP string `ldap:"mcastSrcIP"`
	LVSSyncDaemonInterface string `ldap:"lvsSyncDaemonInterface"`
	VirtualRouterID int `ldap:"virtualRouterID"`
	Priority int `ldap:"priority"`
	AdvertInt int `ldap:"advertInt"`
}


func GenerateDefault(path string, data string, perm string) error {

	// create file
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	// set permissions
	if perm == "" {
		perm = "0600"
	}

	cmd := exec.Command( "chmod", perm, path )
  cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	// write file
	_, err = f.WriteString(data)
	defer f.Close()
	if err != nil {
		return err
	}

	return nil
}

func GenerateKeepalived(inter any) error {


	f := inter.(Kalived)
	log.Println(f)
	return nil
	
}
