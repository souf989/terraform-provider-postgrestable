---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "postgrestable Provider"
subcategory: ""
description: |-
  
---

# postgrestable Provider





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **connect_timeout** (Number) Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
- **database** (String) The name of the database to connect to in order to conenct to (defaults to `postgres`).
- **host** (String) Name of PostgreSQL server address to connect to
- **max_connections** (Number) Maximum number of connections to establish to the database. Zero means unlimited.
- **password** (String, Sensitive) Password to be used if the PostgreSQL server demands password authentication
- **port** (Number) The PostgreSQL port number to connect to at the server host, or socket file name extension for Unix-domain connections
- **scheme** (String)
- **sslmode** (String) This option determines whether or with what priority a secure SSL TCP/IP connection will be negotiated with the PostgreSQL server
- **username** (String) PostgreSQL user name to connect as
