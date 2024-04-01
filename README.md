# :card_index: BUREAU: Centralized Configuration Agent

A lightweight agent for syncing configuration files from LDAP. Includes custom schemas for supported services as well as a default catchall schema.

<!-- asciicinema -->

# Configuration
Bureau intelligently finds and updates relevant config files from your LDAP server, so minimal configuration is necessary.

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

# v1.0 Roadmap
- [ ] Additional schemas (ssh, system, sssd)
- [ ] Kubernetes (configMap) support
- [ ] Secrets management

Raise an issue to request any other services you want to see supported.
