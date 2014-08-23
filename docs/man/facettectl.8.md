% FACETTECTL(8) facette
% Vincent Batoufflet <vincent@batoufflet.info>, Marc Falzon <marc@falzon.me>
% August 23, 2014

# NAME

facettectl - Facette administration utility

# SYNOPSYS

facettectl [*options*] command

# DESCRIPTION

Facette is a a web application to display time series data from various sources — such as collectd, Graphite or
InfluxDB — on graphs.

This utility administrates the application.

# COMMANDS

reload
:   Reload configuration and refresh both catalog and library.

# OPTIONS

-c *file*
:   Specify the application configuration file path (type: string, default: /etc/facette/facette.json).

-h
:   Display application help and exit.

-V
:   Display the application version and exit.

# SEE ALSO

<https://facette.io/>
