# Configuration — Facette

## General Information

This documentation describes [Facette][0] web front-end configuration.

All the configuration are stored as JSON format as described in the [RFC 4627][1] document.

## Main Configuration

### Server Configuration

The main configuration file is passed using `-c` argument when launching the `facette` server.

Mandatory settings:

 * __bind__: the address and port to listen on (type: `string`)
 * __base_dir__: the base Facette application directory holding static files (type: `string`)
 * __data_dir__: the directory used to store application data (type: `string`)
 * __auth_file__: the file containing authentication accounts (type: `string`)
 * __origin_dir__: the path to the folder containing origin configuration files (type: `string`)

Optional settings:

 * __access_log__: the path to the file to store access logging information (type: `string`, default: `stdout`)
 * __server_log__: the path to the file to store Facette application logging data (type: `string`, default: `stdout`)
 * __url_prefix__: the URL prefix behind which the server is located (type: `string`)

Example:

```javascript
{
    "bind": ":12003",
    "base_dir": "/usr/share/facette",
    "data_dir": "/var/lib/facette",
    "auth_file": "/etc/facette/auth.json",
    "origin_dir": "/etc/facette/origins",
    "access_log": "/var/log/facette/access.log",
    "server_log": "/var/log/facette/server.log"
}
```

### Authentication Configuration

An authentication is required to alter Facette configuration (e.g. graphs creation, resources reload, Read-Write API
access). It uses HTTP Basic authentication as described in the 11.1 section of the [RFC 1945][1] document.

Authentication configuration handling is still pretty basic and will need some additional work and refine in the future.
Currently it only stores login and password pairs in a single file.

To create a new user please use the `facettectl` utility (note that you will need to **reload the server** to take into
account the change, use `facettectl reload`):

```
facettectl useradd facette
```

Example (password being `facette'):

```javascript
{
    "facette": "DEnfb3dCCY/NsuET4TwvX8ojD8fhxrcagGd1lbeXqL0="
}
```

## Origins Configuration

TBD


[0]: http://facette.io/
[1]: http://www.ietf.org/rfc/rfc4627.txt
[2]: http://www.ietf.org/rfc/rfc1945.txt
