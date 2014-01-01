# Installation — Facette

## From binaries

Not available yet

## From sources

### Requirements

 * RRD library Go binding: [rrd][0] (along with librrd library and development files)
 * Set package: [set][1]
 * UUID Go package: [gouuid][2]
 * Gorilla [mux][3] and [handlers][4] packages
 * Gopass package: [gopass][5]
 * Stoppable net/http listener package: [stoppableListener][6]

### Build instructions

Retrieve the source code:

```
git clone https://github.com/facette/facette
```

Run the building command:

```
cd facette
make
./build/facette/bin/facette -c path/to/config.json
```

By default Facette will be built and installed in the `build` folder. To change its location use the `PREFIX` variable:

```
PREFIX=/path/to/folder make
/path/to/folder/bin/facette -c path/to/config.json
```

### Additional targets

Run the various test suites:

```
make test
```

Clean the building environment:

```
make clean
```

Note: the `PREFIX` variable must be prepended to each command if passed during the building process.


[0]: https://github.com/ziutek/rrd
[1]: https://github.com/fatih/set
[2]: https://github.com/nu7hatch/gouuid
[3]: https://github.com/gorilla/mux
[4]: https://github.com/gorilla/handlers
[5]: https://github.com/howeyc/gopass
[6]: https://github.com/etix/stoppableListener
