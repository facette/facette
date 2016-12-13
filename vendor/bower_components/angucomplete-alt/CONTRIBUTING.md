## How to Contribute to angucomplete-alt

* Before sending a PR for a feature or bug fix, be sure to run tests by running

```bash
grunt # no arguments, just grunt
```

(If you don't have grunt installed, you'll need to run ``npm install -g grunt-cli`` to install grunt.
You'll also want to run ``bower install && npm install``.)

* If PR is not trivial, please add tests.

* All pull requests should be made to the `master` branch.

* No tabs please. Indent with 2 spaces.

* Do not generate minified version.

* Do not update package.json and bower.json unless you have a strong reason to do it.

### How to run examples:

```bash
cd example
python -m SimpleHTTPServer
```

Open your browser and access http://localhost:8000

