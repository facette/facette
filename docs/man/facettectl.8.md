% FACETTECTL(8) facette
% Vincent Batoufflet <vincent@batoufflet.info>
% May 24, 2014

# NAME

facettectl - Facette administration utility

# SYNOPSYS

facettectl [*options*] command

# DESCRIPTION

Facette is a graphing web front-end for RRD files. This utility administrates the application.

# COMMANDS

reload
:   Reload configuration and refresh both catalog and library.

# OPTIONS

-c *file*
:   Specify the application configuration file path (type: string, default: /etc/facette/facette.json).

-d *level*
:   Specify the server debugging information level (type: integer, default: 0).

# SEE ALSO

<https://facette.io/>
