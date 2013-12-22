% FACETTE(1) facette
% Vincent Batoufflet <vincent@batoufflet.info>
% December 22, 2013

# NAME

facette - graphing web front-end

# SYNOPSYS

facette [*options*]

# DESCRIPTION

Facette is a graphing web front-end for RRD files.

# OPTIONS

-c *file*
:   Specify the application configuration file path (type: string, default: /etc/facette/facette.json).

-d *level*
:   Specify the server debugging information level (type: integer, default: 0).

# SIGNALS

**facette** accepts the following signals:

SIGINT, SIGTERM\
:   These signals cause **facette** to terminate.

SIGHUP\
:   This signal causes **facette** to reload its configuration and to refresh the catalog.

# SEE ALSO

facettectl(8),
<http://facette.io/>
