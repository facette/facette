Facette [![](https://api.travis-ci.org/facette/facette.png)](https://travis-ci.org/facette/facette)
=======

![logo](https://cloud.githubusercontent.com/assets/1122379/3501756/07726d40-061a-11e4-8ffa-bbaa6cf3adfb.png)

What is Facette?
----------------

[Facette][0] is a web application to display time series data from various sources — such as [collectd][1],
[Graphite][2], [InfluxDB][5] or [KairosDB][6] — on graphs, designed to be easy to setup and to use. To learn more on
its architecture, read [this page](http://docs.facette.io/architecture/).

The source code is available at [Github][3] and licensed under the terms of the [BSD license][4].

![facette_sshot2](https://cloud.githubusercontent.com/assets/1122379/3489453/3a61f74e-052e-11e4-884e-ea781b93efdd.png)
![facette_sshot1](https://cloud.githubusercontent.com/assets/1122379/3489442/74b3b000-052d-11e4-812e-e462b8048ebd.png)

Installation
------------

:warning: Facette is currently under development and is **not ready** for a production environment.

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
