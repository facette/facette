module.exports = (grunt)->
  grunt.loadNpmTasks 'grunt-contrib-coffee'
  grunt.loadNpmTasks 'grunt-contrib-watch'
  grunt.loadNpmTasks 'grunt-karma'
  grunt.loadNpmTasks 'grunt-contrib-uglify'
  grunt.loadNpmTasks 'grunt-contrib-copy'

  grunt.initConfig
    coffee:
      compile:
        options:
          sourceMap: true
          watch: true
        files: [
          expand: true
          cwd: 'src/'
          src: '**/*.coffee'
          dest: 'js/'
          ext: '.js'
        ]
    watch:
      source:
        files: [ 'src/**/*' ]
        tasks: [ 'coffee' ]

    karma: 
      unit:
        configFile: 'karma.conf.coffee'
      single:
        configFile: 'karma.conf.coffee'
        options:
          singleRun: true

    uglify:
      source:
        expand: true
        cwd: 'js/'
        src: '**/*.js'
        dest: 'dist/'
        ext: '.min.js'

    copy:
      source:
        expand: true
        cwd: 'js/'
        src: '**/*.js'
        dest: 'dist/'



  grunt.registerTask 'dev', [ 'coffee', 'watch' ]
  grunt.registerTask 'build', [ 'karma:single', 'coffee', 'copy', 'uglify' ]
  grunt.registerTask 'default', [ 'dev' ]
