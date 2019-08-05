Installation
============

Requirements
------------

* Go language environment (>= 1.10)
* RRDtool library and development files (>= 1.4.0)
* pkg-config helper tool
* Node.js and "yarn" package manager
* Pandoc document converter

Debian/Ubuntu:

    apt-get install build-essential golang-go librrd-dev npm nodejs-legacy \
        pandoc pkg-config

Mac OS X (with brew):

    brew install go rrdtool npm pandoc pkg-config

Build Instructions
------------------

Run the building command:

    cd facette
    make
    make install

By default Facette will be built in the `build` directory and installed in the
`/usr/local` one. To change the installation directory set the `PREFIX`
variable:

    sudo make PREFIX=/path/to/directory install

Configuration
-------------

Once installed, follow the configuration steps described here:
http://docs.facette.io/latest/configuration/

Additional Targets
------------------

Run the various test suites:

    make test

Clean the building environment:

    make clean
