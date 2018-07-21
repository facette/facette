"use strict";

var fs = require('fs'),
    gulp = require('gulp'),
    autoprefixer = require('gulp-autoprefixer'),
    chmod = require('gulp-chmod'),
    concat = require('gulp-concat'),
    environments = require('gulp-environments'),
    footer = require('gulp-footer'),
    header = require('gulp-header'),
    htmlmin = require('gulp-htmlmin'),
    jscs = require('gulp-jscs'),
    jshint = require('gulp-jshint'),
    jsonminify = require('gulp-jsonminify'),
    merge = require('merge-stream'),
    path = require('path'),
    rev = require('gulp-rev'),
    revDelete = require('gulp-rev-delete-original'),
    revReplace = require('gulp-rev-replace'),
    templatecache = require('gulp-angular-templatecache'),
    translateextract = require('gulp-angular-translate-extract'),
    uglify = require('gulp-uglify'),
    uglifycss = require('gulp-uglifycss'),
    vendor = require('gulp-concat-vendor');

var config = {
    pkg: JSON.parse(fs.readFileSync('./package.json')),
    banner:
        '/*!\n' +
        ' * <%= pkg.name %> - <%= pkg.description %>\n' +
        ' * Website: <%= pkg.homepage %>\n' +
        ' * License: <%= pkg.license %>\n' +
        ' */\n',
    dist_dir: path.resolve(__dirname, "../dist"),
    files: {
        script: [
            'src/js/extend.js',
            'src/js/chart/chart.js',
            'src/js/chart/config.js',
            'src/js/chart/data.js',
            'src/js/chart/svg.js',
            'src/js/chart/rect.js',
            'src/js/chart/main.js',
            'src/js/chart/title.js',
            'src/js/chart/axis.js',
            'src/js/chart/area.js',
            'src/js/chart/series.js',
            'src/js/chart/tooltip.js',
            'src/js/chart/legend.js',
            'src/js/chart/utils.js',
            'src/js/define.js',
            'src/js/locales.js',
            'src/js/utils.js',
            'src/js/app.js',
            'src/js/api.js',
            'src/js/storage.js',
            'src/js/ui/*.js',
            'src/js/error.js',
            'src/js/common/*.js',
            'src/js/browse/*.js',
            'src/js/show/*.js',
            'src/js/admin/*.js'
        ],
        style: [
            'src/css/font.css',
            'src/css/common.css',
            'src/css/dialog.css',
            'src/css/header.css',
            'src/css/sidebar.css',
            'src/css/content.css',
            'src/css/notify.css',
            'src/css/tab.css',
            'src/css/message.css',
            'src/css/column.css',
            'src/css/list.css',
            'src/css/pagination.css',
            'src/css/sortable.css',
            'src/css/form.css',
            'src/css/menu.css',
            'src/css/tooltip.css',
            'src/css/hotkeys.css',
            'src/css/graph.css'
        ],
        style_print: [
            'src/css/print.css'
        ],
        html: [
            'src/html/ui/*.html',
            'src/html/error/*.html',
            'src/html/admin/*.html',
            'src/html/browse/*.html',
            'src/html/common/*.html',
            'src/html/show/*.html'
        ],
        vendor: {
            js: [
                'node_modules/jquery/dist/jquery.js',
                'node_modules/messageformat/messageformat.js',
                'node_modules/moment/moment.js',
                'node_modules/d3/build/d3.js',
                'node_modules/angular/angular.js',
                'node_modules/angular-route/angular-route.js',
                'node_modules/angular-resource/angular-resource.js',
                'node_modules/angular-sanitize/angular-sanitize.js',
                'node_modules/angular-translate/dist/angular-translate.js',
                'node_modules/angular-translate-loader-static-files/angular-translate-loader-static-files.js',
                'node_modules/angular-translate-interpolation-messageformat/' +
                    'angular-translate-interpolation-messageformat.js',
                'node_modules/angular-inview/angular-inview.js',
                'node_modules/angular_page_visibility/dist/page_visibility.js',
                'node_modules/ng-dialog/js/ngDialog.js',
                'node_modules/ui-select/dist/select.js',
                'node_modules/angular-ui-tree/dist/angular-ui-tree.js',
                'node_modules/angular-paging/dist/paging.js',
                'node_modules/ng-sortable/dist/ng-sortable.js',
                'node_modules/angular-tooltips/dist/angular-tooltips.js',
                'node_modules/angular-bootstrap-colorpicker/js/bootstrap-colorpicker-module.js',
                'node_modules/angularjs-bootstrap-datetimepicker/src/js/datetimepicker.js',
                'node_modules/angular-hotkeys/build/hotkeys.js',
                'node_modules/angular-date-time-input/src/dateTimeInput.js'
            ],
            css: [
                'node_modules/font-awesome/css/font-awesome.min.css'
            ],
            fonts: [
                'node_modules/font-awesome/fonts/fontawesome-webfont.eot',
                'node_modules/font-awesome/fonts/fontawesome-webfont.svg',
                'node_modules/font-awesome/fonts/fontawesome-webfont.ttf',
                'node_modules/font-awesome/fonts/fontawesome-webfont.woff',
                'node_modules/font-awesome/fonts/fontawesome-webfont.woff2',
                'node_modules/typeface-roboto/files/roboto-latin-300.woff',
                'node_modules/typeface-roboto/files/roboto-latin-300.woff2',
                'node_modules/typeface-roboto/files/roboto-latin-400.woff',
                'node_modules/typeface-roboto/files/roboto-latin-400.woff2',
                'node_modules/typeface-roboto/files/roboto-latin-500.woff',
                'node_modules/typeface-roboto/files/roboto-latin-500.woff2'
            ],
            images: [
                'src/images/*'
            ]
        }
    }
};

var buildTasks = [
    'build-scripts',
    'build-styles',
    'build-html',
    'build-locales',
    'copy-styles',
    'copy-html'
];

gulp.task('default', [
    'build'
]);

gulp.task('build', environments.production() ? buildTasks.concat('rev-replace') : buildTasks);

gulp.task('lint', [
    'lint-scripts'
]);

gulp.task('build-scripts', ['build-html'], function() {
    return merge(
        gulp.src(config.files.script.concat([config.dist_dir + '/tmp/templates.js']))
            .pipe(concat('facette.js'))
            .pipe(header(config.banner + '\n(function() {\n\n"use strict";\n\n', {pkg: config.pkg}))
            .pipe(footer('\n}());\n'))
            .pipe(environments.production(uglify({mangle: false, preserveComments: 'license'})))
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/js')),

        gulp.src(config.files.vendor.js)
            .pipe(vendor('vendor.js'))
            .pipe(environments.production(uglify({preserveComments: 'license'})))
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/js'))
    );
});

gulp.task('lint-scripts', function() {
    return gulp.src(config.files.script)
        .pipe(jshint())
        .pipe(jshint.reporter())
        .pipe(jscs())
        .pipe(jscs.reporter());
});

gulp.task('build-styles',function() {
    return merge(
        gulp.src(config.files.style)
            .pipe(concat('style.css'))
            .pipe(header(config.banner + '\n', {pkg: config.pkg}))
            .pipe(autoprefixer())
            .pipe(environments.production(uglifycss()))
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/css')),

        gulp.src(config.files.style_print)
            .pipe(concat('style-print.css'))
            .pipe(header(config.banner + '\n', {pkg: config.pkg}))
            .pipe(autoprefixer())
            .pipe(environments.production(uglifycss()))
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/css'))
    );
});

gulp.task('copy-styles', function() {
    return merge(
        gulp.src(config.files.vendor.css)
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/css')),

        gulp.src(config.files.vendor.fonts)
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/fonts')),

        gulp.src(config.files.vendor.images)
            .pipe(chmod(644))
            .pipe(gulp.dest(config.dist_dir + '/assets/images'))
    );
});

gulp.task('build-html', function() {
    return gulp.src(config.files.html)
        .pipe(htmlmin({collapseWhitespace: true}))
        .pipe(templatecache({
            base: process.cwd(),
            module: 'facette',
            transformUrl: function(url) {
                url = url.substr(9);
                if (url.indexOf('ui/') === 0) {
                    url = url.substr(3);
                }

                return 'templates/' + url;
            }
        }))
        .pipe(gulp.dest(config.dist_dir + '/tmp'));
});

gulp.task('copy-html', function() {
    return gulp.src('src/html/index.html')
        .pipe(chmod(644))
        .pipe(gulp.dest(config.dist_dir + '/assets/html'));
});

gulp.task('build-locales', function() {
    return gulp.src('src/js/locales/*.json')
        .pipe(jsonminify())
        .pipe(chmod(644))
        .pipe(gulp.dest(config.dist_dir + '/assets/js/locales'));
});

gulp.task('update-locales', function() {
    return gulp.src(config.files.script.concat(['src/html/index.html', config.files.html]))
        .pipe(translateextract({
            lang: ['en', 'fr'],
            suffix: '.json',
            dest: 'src/js/locales',
            nullEmpty: true,
            safeMode: true,
            stringifyOptions: true
        }))
        .pipe(gulp.dest('src/js'));
});

gulp.task('rev-rename', buildTasks, function() {
    return gulp.src(config.dist_dir + '/assets/{css,fonts,images,js}/*', {base: config.dist_dir + '/assets'})
        .pipe(rev())
        .pipe(revDelete())
        .pipe(gulp.dest(config.dist_dir + '/assets'))
        .pipe(rev.manifest('rev-manifest.json'))
        .pipe(gulp.dest(config.dist_dir + '/tmp'));
});

gulp.task('rev-replace', ['rev-rename'], function() {
    return gulp.src(config.dist_dir + '/assets/*/*', {base: config.dist_dir + '/assets'})
        .pipe(revReplace({manifest: gulp.src(config.dist_dir + '/tmp/rev-manifest.json')}))
        .pipe(gulp.dest(config.dist_dir + '/assets'));
});
