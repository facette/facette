Facette [![](https://api.travis-ci.org/facette/facette.svg?branch=master)](https://travis-ci.org/facette/facette)
=======

[Facette][0] is a open source web application to display time series data from various sources — such as [collectd][1],
[Graphite][2], [InfluxDB][5] or [KairosDB][6] — on graphs. To learn more on its architecture, read
[this page](http://docs.facette.io/latest/architecture/).

The source code is available at [Github][3] and is licensed under the terms of the [BSD license][4].

![](https://facette.io/assets/images/sshot-view1.png)

Installation
------------

Please see [INSTALL.md](INSTALL.md) file for build instructions and installation procedures.

Contribution
------------

We welcome all your contributions. So, don't hesitate to fork the project, make your changes and submit us your pull
requests.

However, as Facette is under development and still subject to heavy changes, please open an issue to discuss yours if
you think that they will have quite an impact on the code base before starting contributing.

To make the things easier, we will ask for the following:

 * Always use `go fmt`
 * Keep code lines length under 120 characters
 * Provide (when applicable) unit tests for the new code
 * Make sure to run `make test`, having the process completing successfully
 * Squash your commits into a single commit


[0]: https://facette.io/
[1]: https://collectd.org/
[2]: https://graphite.readthedocs.org/
[3]: https://github.com/facette/facette
[4]: https://opensource.org/licenses/BSD-3-Clause
[5]: https://influxdb.com/
[6]: https://kairosdb.github.io/
