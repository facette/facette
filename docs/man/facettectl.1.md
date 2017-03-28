% FACETTECTL(1) facette
% Vincent Batoufflet <vincent@batoufflet.info>, Marc Falzon <marc@falzon.me>
% March 28, 2017

# NAME

Facette - Time series data visualization software

# SYNOPSYS

facettectl [*options*] *command* [*args*...]

# DESCRIPTION

Facette control utility.

# OPTIONS

-a, --address=*http://localhost:12003*
:   Set upstream socket address.

-t, --timeout=*30*
:   Set upstream connection timeout.

-h, --help
:   Display application help and exit.

-v, --verbose
:   Run in verbose mode.

# COMMANDS

help [*command*...]
:   Display command help.

version
:   Display version and support information.

library dump [*options*]
:   Dump data from library.

    -o, --output=*path*  Set dump output file path.

library restore --input=*path* [*options*]
:   Restore data from dump into library.

    -i, --input=*path*  Set dump input file path.
    -m, --merge         Merge data with existing library.

# SEE ALSO

facette(1),
<https://facette.io/>
