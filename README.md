Facette [![Travis CI][travis-badge]][travis-url]
=======

[Facette][project-url] is a open source web application to display time
series data from various sources — such as [collectd][collectd-url],
[Graphite][graphite-url], [InfluxDB][influxdb-url] or
[KairosDB][kairosdb-url] — on graphs. To learn more on its architecture,
read [this page][project-arch].

The source code is available at [Github][project-source] and is licensed
under the terms of the [BSD license][project-license].

![Screenshot][project-sshot]

Installation
------------

Please see [INSTALL.md](INSTALL.md) file for build instructions and
installation procedures.

Contribution
------------

We welcome all your contributions. So, don't hesitate to fork the project,
make your changes and submit us your pull requests.

However, as Facette is under development and still subject to heavy changes,
please open an issue to discuss yours if you think that they will have quite
an impact on the code base before starting contributing.

To make the things easier, we will ask for the following:

* Always use `gofmt`
* Keep code lines length under 120 characters
* Provide (when applicable) unit tests for the new code
* Make sure to run `make test`, having the process completing successfully
* Squash your commits into a single commit

[collectd-url]: https://collectd.org/
[graphite-url]: https://graphite.readthedocs.org/
[influxdb-url]: https://influxdb.com/
[kairosdb-url]: https://kairosdb.github.io/
[project-arch]: https://docs.facette.io/latest/architecture/
[project-license]: https://opensource.org/licenses/BSD-3-Clause
[project-source]: https://github.com/facette/facette
[project-sshot]: https://facette.io/assets/images/sshot-view1.png
[project-url]: https://facette.io/
[travis-badge]: https://api.travis-ci.org/facette/facette.svg?branch=master
[travis-url]: https://travis-ci.org/facette/facette
