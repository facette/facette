% FACETTED(1) facetted
% Vincent Batoufflet <vincent@batoufflet.info>
% November 16, 2013

# NAME

facetted - graphing web front-end

# SYNOPSYS

facetted [*options*] -c file

# DESCRIPTION

Facette is a graphing web front-end for RRD files.

# OPTIONS

-c *file*
:   Specify the application configuration file path (**mandatory**).

-d *level*
:   Specify the server debugging information level (type: integer, default: 0).

# SIGNALS

**facetted** accepts the following signals:

SIGINT, SIGTERM\ 
:   These signals cause **facetted** to terminate.

SIGHUP\ 
:   This signal causes **facetted** to reload its configuration and to refresh the catalog.

# SEE ALSO

<http://facette.io/>
