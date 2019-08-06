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
    tmp = require('tmp'),
    translateextract = require('gulp-angular-translate-extract'),
    terser = require('gulp-terser'),
    uglifycss = require('gulp-uglifycss');

var dist_dir = path.resolve(__dirname, '../dist'),
    tmp_dir = tmp.dirSync({unsafeCleanup: true}).name;

var config = {
    pkg: JSON.parse(fs.readFileSync('./package.json')),
    banner:
        '/*!\n' +
        ' * <%= pkg.name %> - <%= pkg.description %>\n' +
        ' * Website: <%= pkg.homepage %>\n' +
        ' * License: <%= pkg.license %>\n' +
        ' */\n',
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
                'node_modules/angularjs-tooltips/dist/angular-tooltips.js',
                'node_modules/angular-bootstrap-colorpicker/js/bootstrap-colorpicker-module.js',
                'node_modules/angularjs-bootstrap-datetimepicker/src/js/datetimepicker.js',
                'node_modules/angular-hotkeys/build/hotkeys.js',
                'node_modules/boula/dist/boula.js',
                'node_modules/angular-date-time-input/src/dateTimeInput.js' // should be the last (vendoring issue)
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

function buildHtml() {
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
        .pipe(gulp.dest(tmp_dir));
}

function buildScripts() {
    return merge(
        gulp.src(config.files.script.concat([tmp_dir + '/templates.js']), { allowEmpty: true })
            .pipe(concat('facette.js'))
            .pipe(header(config.banner + '\n(function() {\n\n"use strict";\n\n', {pkg: config.pkg}))
            .pipe(footer('\n}());\n'))
            .pipe(environments.production(terser({ mangle: false, output: { comments: 'some' } })))
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/js')),

        gulp.src(config.files.vendor.js, { allowEmpty: true })
            .pipe(concat('vendor.js'))
            .pipe(environments.production(terser({ output: { comments: 'some' } })))
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/js'))
    );
}

function buildLocales() {
    return gulp.src('src/js/locales/*.json')
        .pipe(jsonminify())
        .pipe(chmod(0o644))
        .pipe(gulp.dest(dist_dir + '/assets/js/locales'));
}

function buildStyles() {
    return merge(
        gulp.src(config.files.style)
            .pipe(concat('style.css'))
            .pipe(header(config.banner + '\n', {pkg: config.pkg}))
            .pipe(autoprefixer())
            .pipe(environments.production(uglifycss()))
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/css')),

        gulp.src(config.files.style_print)
            .pipe(concat('style-print.css'))
            .pipe(header(config.banner + '\n', {pkg: config.pkg}))
            .pipe(autoprefixer())
            .pipe(environments.production(uglifycss()))
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/css'))
    );
}

var build = gulp.parallel(
    gulp.series(
        buildHtml,
        buildScripts
    ),
    buildLocales,
    buildStyles,
    copyHtml,
    copyStyles,
);
if (environments.production()) {
    build = gulp.series(build, revRenameTask, revReplaceTask);
}

function lintScripts() {
    return gulp.src(config.files.script)
        .pipe(jshint())
        .pipe(jshint.reporter())
        .pipe(jscs())
        .pipe(jscs.reporter());
}

var lint = gulp.series(lintScripts);

function copyHtml() {
    return gulp.src('src/html/index.html')
        .pipe(chmod(0o644))
        .pipe(gulp.dest(dist_dir + '/assets/html'));
}

function copyStyles() {
    return merge(
        gulp.src(config.files.vendor.css, { allowEmpty: true })
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/css')),

        gulp.src(config.files.vendor.fonts)
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/fonts')),

        gulp.src(config.files.vendor.images)
            .pipe(chmod(0o644))
            .pipe(gulp.dest(dist_dir + '/assets/images'))
    );
}

function updateLocales() {
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
}

function revRenameTask() {
    return gulp.src(dist_dir + '/assets/{css,fonts,images,js}/*', {base: dist_dir + '/assets'})
        .pipe(rev())
        .pipe(revDelete())
        .pipe(gulp.dest(dist_dir + '/assets'))
        .pipe(rev.manifest('rev-manifest.json'))
        .pipe(gulp.dest(tmp_dir));
}

function revReplaceTask() {
    return gulp.src(dist_dir + '/assets/*/*', {base: dist_dir + '/assets'})
        .pipe(revReplace({manifest: gulp.src(tmp_dir + '/rev-manifest.json')}))
        .pipe(gulp.dest(dist_dir + '/assets'));
}

exports.buildHtml = buildHtml;
exports.buildLocales = buildLocales;
exports.buildScripts = gulp.series(buildHtml, buildScripts);
exports.buildStyles = buildStyles;
exports.build = build;

exports.lintScripts = lintScripts;
exports.lint = lint;

exports.copyStyles = copyStyles;
exports.copyHtml = copyHtml;

exports.updateLocales = updateLocales;

exports.revRename = gulp.series(build, revRenameTask);
exports.revReplace = gulp.series(build, revRenameTask, revReplaceTask);

exports.default = build;
