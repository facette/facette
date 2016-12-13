/* jshint node:true */

var nodeModules = [
  'node_modules/jquery/dist/jquery.js',
  'node_modules/moment/moment.js',
  'node_modules/angular/angular.js',
  'node_modules/angular-mocks/angular-mocks.js'
]
var bumpFiles = ['package.json', 'bower.json', 'README.md', 'src/js/*.js']
var miscFiles = ['GruntFile.js', 'gulpfile.js', 'karma.conf.js', 'paths.js']
var demoFiles = []
var sourceFiles = ['src/**/*.js']
var testFiles = ['test/**/*.spec.js']

module.exports = {
  all: nodeModules.concat(sourceFiles).concat(testFiles).concat(demoFiles),
  app: sourceFiles,
  bump: bumpFiles,
  lint: miscFiles.concat(sourceFiles).concat(testFiles).concat(miscFiles),
  src: sourceFiles,
  test: testFiles
}
