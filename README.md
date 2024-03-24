# BUREAU: Centralized Configuration Agent (FIRST RELEASE COMING SOON)

A lightweight agent for syncing configuration files from LDAP. Includes custom schemas for supported services as well as a default catchall scheme.

# Quickstart
1. Install LDAP schemas and add configuration files.

2. Enter LDAP server information into configuration file (`/etc/bureau/bureau.yaml`, `~/.config/bureau/bureau.yaml`, or `~/.bureau.yaml`).

3. Start `bureau` in daemon mode or with systemd.

# Built-in Schemas
- Default
  - Any file, any location
- Keepalived
  - Global settings
  - Sync group
  - VRRP instance

# Full Example Setup


# Coming Soon
- [ ] sssd
- [ ] ssh
- [ ] systemd
- [ ] Kubernetes (configMap) support
- [ ] secrets management

Raise an issue to request any other services you want to see supported.
