% FACETTE(1) facette
% Vincent Batoufflet <vincent@batoufflet.info>, Marc Falzon <marc@falzon.me>
% August 23, 2014

# NAME

Facette - Time series data visualization and graphing software

# SYNOPSYS

facette [*options*]

# DESCRIPTION

Facette is a a web application to display time series data from various sources — such as collectd, Graphite or
InfluxDB — on graphs.

# OPTIONS

-c *file*
:   Specify the application configuration file path (type: string, default: /etc/facette/facette.json).

-h
:   Display application help and exit.

-l *file*
:   Specify the server log file (type: string, default: STDERR)

-L *level*
:   Specify the server logging level (type: string, default: info).

    Supported levels: error, warning, notice, info, debug.

-V
:   Display the application version and exit.

# SIGNALS

**facette** accepts the following signals:

SIGINT, SIGTERM
:   These signals cause **facette** to terminate.

SIGUSR1
:   This signal causes **facette** to refresh its catalog and library.

# SEE ALSO

facettectl(8),
<https://facette.io/>
