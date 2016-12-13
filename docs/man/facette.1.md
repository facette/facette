% FACETTE(1) facette
% Vincent Batoufflet <vincent@batoufflet.info>, Marc Falzon <marc@falzon.me>
% December 13, 2016

# NAME

Facette - Time series data visualization software

# SYNOPSYS

facette [*options*]

# DESCRIPTION

Facette is a web application to display time series data from various sources — such as collectd, Graphite, InfluxDB
or KairosDB.

# OPTIONS

-c *file*
:   Specify the application configuration file path (default: /etc/facette/facette.conf).

-h
:   Display application help and exit.

-V
:   Display the application version and exit.

# SIGNALS

**facette** accepts the following signals:

SIGINT, SIGQUIT, SIGTERM
:   These signals cause **facette** to terminate.

SIGUSR1
:   This signal causes **facette** to refresh its catalog and library.

# SEE ALSO

facettectl(8),
<https://facette.io/>
