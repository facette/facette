% FACETTE(1) facette
% Vincent Batoufflet <vincent@batoufflet.info>, Marc Falzon <marc@falzon.me>
% August 17, 2018

# NAME

Facette - Time series data visualization software

# SYNOPSYS

facette [*options*]

# DESCRIPTION

Facette is a web application to display time series data from various sources
— such as collectd, Graphite, InfluxDB or KairosDB.

# OPTIONS

-c, --config=*/etc/facette/facette.yaml*
:   Specify the application configuration file path.

-h, --help
:   Display application help and exit.

-V, --version
:   Display the application version and exit.

# SIGNALS

**facette** accepts the following signals:

SIGINT, SIGQUIT, SIGTERM
:   These signals cause **facette** to terminate.

SIGUSR1
:   This signal causes **facette** to refresh its catalog.

# SEE ALSO

facettectl(1),
<https://facette.io/>
