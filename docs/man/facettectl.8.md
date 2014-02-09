% FACETTECTL(8) facette
% Vincent Batoufflet <vincent@batoufflet.info>
% December 22, 2013

# NAME

facettectl - Facette administration utility

# SYNOPSYS

facettectl [*options*] command

# DESCRIPTION

Facette is a graphing web front-end for RRD files. This utility administrates the application.

# COMMANDS

useradd *name*
:   Create a new user into the authentication backend.

userdel *name*
:   Remove an existing user from the authentication backend.

userlist
:   List all the existing user entries.

usermod *name*
:   Modify an existing user.

# OPTIONS

-c *file*
:   Specify the application configuration file path (type: string, default: /etc/facette/facette.json).

-d *level*
:   Specify the server debugging information level (type: integer, default: 0).

# SEE ALSO

<http://facette.io/>
