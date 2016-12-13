module.exports = function(grunt) {

  // Project configuration.
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),
    uglify: {
      options: {
        banner: '/*! ' + 
                '\n* <%= pkg.name %> v<%= pkg.version %> by <%= pkg.author %> - <%= pkg.license %> licensed ' +
                '\n* <%= pkg.repository.url %> ' + 
                '\n*/\n'
      },
      build: {
        src: 'src/<%= pkg.distName %>.js',
        dest: 'dist/<%= pkg.distName %>.min.js'
      }
    },
    copy: {
      main: {
        src: 'src/<%= pkg.distName %>.js',
        dest: 'dist/<%= pkg.distName %>.js'
      }
    }
  });

  // Load the plugin that provides the "uglify" task.
  grunt.loadNpmTasks('grunt-contrib-uglify');
  
  // Load the plugin that provides the "copy" task
  grunt.loadNpmTasks('grunt-contrib-copy')

  // Default task(s).
  grunt.registerTask('default', ['uglify', 'copy']);

};