# BUREAU: Centralized Configuration Agent

A lightweight agent for syncing any and all configuration files with LDAP. Includes custom schemas for supported services as well as a default catchall schema.

# Currently Supported
- Default
  - Any file, any location
- Keepalived
  - Global settings
  - Sync group
  - VRRP instance

# Coming Soon
- apache2
- sssd
- Kubernetes (configMap) support

Raise an issue to request other supported services.
