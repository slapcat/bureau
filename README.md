# :card_index: BUREAU: Centralized Configuration Agent

Bureau is a lightweight agent designed to synchronize configuration files from an LDAP server. It supports custom schemas for specific services and includes a default catchall schema.

# Features

- **RFC-Compliant:** Works with standard LDAP servers (OpenLDAP, Active Directory, 389, etc.).
- **Customizable:**  Uses globally valid OIDs and modifiable templates for application-specific configurations.
- **Versatile:** Practical for both servers and desktop users.
- **Fast & Efficient:** Lightweight design with in-memory tracking of LDAP changes.

# Quickstart
1. Add the schemas to your LDAP directory:
```
git clone https://github.com/slapcat/bureau.git
ldapadd -Y EXTERNAL -H ldapi:/// -f bureau/schemas/configFile.ldif
ldapadd -Y EXTERNAL -H ldapi:/// -f bureau/schemas/keepalivedGlobalConfig.ldif
ldapadd -Y EXTERNAL -H ldapi:/// -f bureau/schemas/keepalivedVRRPGroupConfig.ldif
ldapadd -Y EXTERNAL -H ldapi:/// -f bureau/schemas/keepalivedVRRPInstanceConfig.ldif
```

2. Create a test config file:
```
ldapadd -Y EXTERNAL -H ldapi:/// <<EOF
dn: cn=bureau,cn=<hostname>,ou=config,dc=example,dc=com
path: /tmp/bureau.txt
cn: bureau
data: Hello World!
objectClass: configFile
objectClass: top
EOF
```

3. Install bureau on your target system and add the LDAP server credentials to `bureau.yaml`.

4. Start bureau in daemon mode or with systemd:
```
./bureau &                               # daemon
systemctl enable --now bureau.service    # systemd
```

Systemd will generate files owned by `root:root`. If you want to use bureau for user files, you can copy the systemd unit file to the user-specific directory:
```
cp /usr/lib/systemd/system/bureau.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl enable --user --now bureau.service
```

5. The new file should be available instantly:
```
$ cat /tmp/bureau.txt
Hello world!
```

# Configuration

Bureau requires minimal setup, intelligently finding and updating relevant config files.

**bureau.yaml**
```yaml
debug: true
daemon: true
update_interval: 600
restart_service_on_change: true

server: ldap://ldap.example.com
binddn: cn=bureau,ou=services,dc=example,dc=com
password: SomeSecretPassword
base: ou=config,dc=example,dc=com
host_specific_entries: true
override_hostname:
```

**Location**

The bureau configuration is looked for in these locations in order of precedence:
- same directory as the binary
- ~/.bureau.yaml
- ~/.config/bureau/bureau.yaml
- /etc/bureau/bureau.yaml

**Daemon Mode**

`daemon` mode will run the bureau binary as a service. This mode is the default in order to benefit from bureau's in-memory tracking of LDAP changes, which will only pull entire entries if they have a more recent `modifyTimestamp` than the previous time it was checked. It is recommended to use daemon mode in addition to the systemd service. `update_interval` specifies the number of seconds between each LDAP search for new config files.

**Host Specific Entries**

This settings indicates that all relevant config files for the current host are stored under `cn=<hostname>,<base>`, for example `cn=web01,ou=config,dc=example,dc=com`. The entries under this DN can be grouped or named however you wish. If you disable this option, bureau will sync all config files listed under the base DN.

In some cases it is useful to specify an identical set of config files for multiple systems. This can be achieved by setting an `override_hostname`. This option designates the CommonName used when searching for config files. For example, setting `override_hostname` to `desktop` will search for config files under `cn=desktop,ou=config,dc=example,dc=com`.

# Built-in Schemas
- configFile
  - Any file, any location
- Keepalived-specific
  - keepalivedGlobalConfig
  - keepalivedVRRPGroupConfig
  - keepalivedVRRPInstanceConfig
