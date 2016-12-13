/* globals require, __dirname */
/* jshint node:true */
'use strict'

var gulp = require('gulp')
var lodash = require('lodash')
var path = require('path')
var paths = require('./paths')
var plato = require('plato')
var Server = require('karma').Server
var standard = require('gulp-standard')

var karmaConfig = path.join(__dirname, 'karma.conf.js')

gulp.task('clean', function () {
  var del = require('del')
  return del([
    'build'
  ])
})

gulp.task('complexity', function (done) {
  var callback = function () {
    done()
  }

  plato.inspect(paths.lint, 'build/complexity', {title: 'prerender', recurse: true}, callback)
})

gulp.task('lint', function () {
  return gulp
    .src(paths.lint)
    .pipe(standard())
    .pipe(standard.reporter('default', {
      breakOnError: true,
      quiet: true
    }))
})

gulp.task('tdd', function (done) {
  gulp.watch(paths.all, ['jscs', 'lint'])

  var config = testConfig(
    {
      autoWatch: true,
      browsers: ['PhantomJS'],
      configFile: karmaConfig,
      singleRun: false
    }
  )

  var server = new Server(config, done)
  server.start()
})

gulp.task('test', ['lint'], function (done) {
  var config = testConfig(
    {
      configFile: karmaConfig,
      singleRun: true,
      reporters: ['progress', 'coverage', 'threshold']
    }
  )

  var server = new Server(config, done)
  server.start()
})

gulp.task('default', ['complexity', 'test'])

var testConfig = function (options) {
  var travisDefaultOptions = {
    browsers: ['Firefox'],
    reporters: ['dots', 'coverage', 'threshold']
  }

  var travisOptions = process.env.TRAVIS && travisDefaultOptions

  return lodash.assign(options, travisOptions)
}
