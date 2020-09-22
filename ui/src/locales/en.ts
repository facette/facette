/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

export default {
    name: "English",

    date: {
        long: "MMM d, yyyy, HH:mm:ss",
    },

    help: {
        charts: {
            axes: {
                label: "Label of the axis. @:help.common.templateSupport",
                max: "Maximum value of the axis.",
                min: "Minimum value of the axis.",
            },
            name: "Name of the chart. @:help.common.name",
            title: "Title of the chart. @:help.common.templateSupport",
        },
        common: {
            name:
                "Must start and end by an alphanumerical character, and contain alphanumerical characters, hyphens " +
                "or underscores.",
            templateSupport: "This field supports template variables.",
        },
        dashboards: {
            name: "Name of the dashboard. @:help.common.name",
            title: "Title of the dashboard. @:help.common.templateSupport",
        },
        database: {
            archive:
                "You can download an archive dumping all objects from the back-end storage database for backup " +
                "and restore purposes.",
            restoreWarning: "Please note that all previously existing data will be overwritten upon restore.",
        },
        filters: {
            action:
                "Action to be performed by the filter:\n" +
                "* `discard`: drops matching metrics\n" +
                "* `relabel`: relabels matching metrics\n" +
                "* `rewrite`: rewrites label of matching metrics\n" +
                "* `sieve:` keeps only matching metrics\n",
            into: "Replacement value to apply on the value associated with the filter label.",
            label: "Label on which the filter will be applied.",
            pattern: "Pattern to apply on the value associated with the filter label. Must follow the RE2 syntax.",
        },
        keyboard: {
            shortcuts:
                "Save time navigating Facette by using keyboard shortcuts.\n\n" +
                "**Note:** Learn more about available shortcuts by visiting their dedicated help section.",
        },
        providers: {
            name: "Name of the provider. @:help.common.name",
            pollInterval: "Provider metrics automatic polling interval. Disabled if empty.",
            prometheus: {
                filter: "Filter for querying metrics from upstream Prometheus service.",
            },
            rrdtool: {
                path: "Base directory from which to search for files.",
                pattern: "Pattern to apply to found files paths. Must follow the RE2 syntax.",
                daemon: "rrdcached daemon socket address.",
            },
            url: "URL to the upstream {0} service.",
        },
        refresh: {
            interval: "Time interval for automatic refresh in seconds. Use either empty or `0` to disable.",
        },
    },

    labels: {
        adminPanel: "Administration panel",
        basket: {
            _: "Basket",
            add: "Add to basket",
            clear: "Clear basket",
            preview: "Preview basket…",
            refresh: "Refresh basket",
            remove: "Remove from basket",
        },
        cancel: "Cancel",
        catalog: "Catalog",
        charts: {
            _: "Chart | Charts",
            axes: {
                _: "Axis | Axes",
                left: "Left",
                max: "Max",
                min: "Min",
                right: "Right",
                select: "Select an axis…",
                x: "X",
                yLeft: "Left Y",
                yRight: "Right Y",
            },
            create: "Create chart",
            delete: "Delete chart | Delete charts",
            edit: "Edit chart",
            legend: {
                _: "Legend",
                show: "Show legend",
            },
            name: "Chart name",
            new: "New chart",
            filter: "Filter charts",
            preview: "Preview chart",
            refresh: "Refresh chart",
            reset: "Reset chart",
            save: "Save chart",
            type: {
                _: "Type",
                area: "Area",
                bar: "Bar",
                line: "Line",
                select: "Select a type…",
            },
            zoom: {
                in: "Zoom in",
                out: "Zoom out",
            },
        },
        clearSelection: "Clear selection",
        clipboard: {
            copy: "Copy to clipboard",
        },
        clone: "Clone",
        close: "Close",
        color: "Color",
        connectors: {
            _: "Connector | Connectors",
            select: "Select a connector…",
        },
        custom: "Custom…",
        dashboards: {
            _: "Dashboard | Dashboards",
            delete: "Delete dashboard | Delete dashboards",
            edit: "Edit dashboard",
            filter: "Filter dashboards",
            name: "Dashboard name",
            new: "New dashboard",
            refresh: "Refresh dashboard",
            reset: "Reset dashboard",
            save: "Save dashboard",
            saveAs: "Save as dashboard…",
        },
        database: {
            _: "Database",
            downloadArchive: "Download archive",
            driver: "{name}, version {version}",
            restore: "Restore",
            restoreArchive: "Restore archive",
        },
        default: "Default",
        delete: "Delete",
        display: "Display",
        displayHelp: "Display this help",
        documentation: "Documentation",
        empty: "Empty",
        export: {
            _: "Export",
            imagePNG: "Save as PNG…",
            summaryCSV: "Summary as CSV…",
            summaryJSON: "Summary as JSON…",
            textMarkdown: "Save as Markdown…",
        },
        expr: {
            none: "No expression",
        },
        filters: {
            _: "Filter | Filters",
            action: {
                _: "Action",
                select: "Select an action…",
            },
            add: "Add filter",
            edit: "Edit filter",
            into: "Into",
            pattern: "Pattern",
            remove: "Remove filter",
            set: "Set filter",
            targets: {
                _: "Targets",
                add: "Add target",
            },
        },
        format: "Format",
        fullscreen: {
            enter: "Enter full screen",
            leave: "Leave full screen",
            toggle: "Toggle full screen",
        },
        general: "General",
        goto: {
            adminPanel: "Go to administration panel",
            charts: "Go to chart | Go to charts",
            chartBack: "Go back to chart",
            dashboards: "Go to dashboard | Go to dashboards",
            dashboardBack: "Go back to dashboard",
            home: "Go to home",
            metrics: "Go to metrics",
            providers: "Go to providers",
        },
        help: "Help",
        home: "Home",
        info: {
            _: "Information",
            branch: "Branch",
            buildDate: "Build date",
            compiler: "Compiler",
            connectors: "Supported connectors",
            revision: "Revision",
            version: "Version",
        },
        items: {
            remove: "Remove item",
            unsupported: "Unsupported item",
        },
        keyboard: {
            _: "Keyboard",
            shortcuts: {
                _: "Keyboard shortcuts",
                enable: "Enable keyboard shortcuts",
            },
        },
        labels: {
            _: "Label | Labels",
            explorer: "Labels explorer",
            search: "Search labels",
        },
        language: {
            _: "Language",
            select: "Select a language…",
        },
        lastModified: "Last modified",
        layout: "Layout",
        leavePage: "Leave page",
        library: "Library",
        markers: {
            _: "Marker | Markers",
            add: "Add marker",
            edit: "Edit marker",
            remove: "Remove marker",
            set: "Set marker",
        },
        metrics: {
            _: "Metric | Metrics",
            fetching: "Fetch metrics…",
            filter: "Filter metrics",
            matching: "Matching {0} metric | Matching {0} metrics",
        },
        moreActions: "More actions…",
        name: {
            _: "Name",
            choose: "Choose a name",
        },
        ok: "OK",
        openMenu: "Open menu",
        options: "Options",
        placeholders: {
            default: "default: {0}",
            example: "e.g. {0}",
        },
        properties: "Properties",
        providers: {
            _: "Provider | Providers",
            delete: "Delete provider | Delete providers",
            disable: "Disable",
            disabled: "Providers is disabled",
            enable: "Enable",
            enabled: "Providers is enabled",
            filter: "Filter providers",
            name: "Provider name",
            new: "New provider",
            poll: "Poll",
            pollAlt: "Poll providers",
            pollInterval: "Poll interval",
            reset: "Reset provider",
            rrdtool: {
                path: "Path",
                pattern: "Pattern",
                daemon: "Daemon address",
            },
            save: "Save provider",
            test: "Test provider",
        },
        refresh: {
            _: "Refresh",
            interval: "Refresh interval",
            list: "Refresh list",
            next: "Next refresh in {0}",
            reset: "Reset interval",
            setInterval: "Set interval",
        },
        reset: "Reset",
        results: "Results",
        retry: "Retry",
        saveAndGo: "Save and Go",
        series: {
            _: "Series | Series",
            add: "Add series",
            edit: "Edit series",
            remove: "Remove series",
            set: "Set series",
        },
        settings: {
            _: "Settings…",
            apply: "Apply settings",
            display: "Display settings",
            personal: "Personal settings",
        },
        show: {
            _: "Show",
            more: "Show more",
        },
        system: "System",
        templates: {
            _: "Template | Templates",
            edit: "Edit template",
            instance: "Template instance",
            newFrom: "New from template",
            save: "Save template",
            select: "Select a template…",
        },
        theme: {
            _: "Theme",
            auto: "Automatic (system preference)",
            select: "Select a theme…",
        },
        timeRange: {
            _: "Time range",
            autoPropagate: "Automatically propagate time range",
            from: "From",
            multiple: "Multiple time ranges",
            propagate: "Propagate time range",
            reset: "Reset time range",
            set: "Set time range",
            to: "To",
            units: {
                days: "Last {count} day | Last {count} days",
                hours: "Last {count} hour | Last {count} hours",
                minutes: "Last {count} minute | Last {count} minutes",
                months: "Last {count} month | Last {count} months",
                years: "Last {count} year | Last {count} years",
            },
        },
        timezone: {
            _: "Time zone",
            local: "Local time",
            select: "Select a time zone…",
            utc: "UTC",
        },
        title: "Title",
        tls: {
            skipVerify: "Skip server certificate verification (Insecure)",
        },
        toggleSidebar: "Toggle sidebar",
        unnamed: "Unnamed",
        url: "URL",
        value: "Value",
        variables: {
            _: "Variables",
            clear: "Clear variable",
            dynamic: "Dynamic",
            edit: "Edit variable",
            set: "Set variable",
            fixed: "Fixed",
        },
        visit: {
            documentation: "Visit documentation",
            website: "Visit website",
        },
    },

    messages: {
        basket: {
            empty: "Basket is empty",
        },
        charts: {
            conflict: "A chart with the same name already exists.",
            delete:
                "You are about to delete the “{name}” chart. Are you sure? | " +
                "You are about to delete {count} charts. Are you sure?",
            deleted: "Chart successfully deleted | Charts successfully deleted",
            none: "No charts defined",
            notFound: "Chart not found",
            saved: "Chart successfully saved",
            selected: "{0} chart selected | {0} charts selected",
        },
        copied: "Copied!",
        dashboards: {
            conflict: "A dashboard with the same name already exists.",
            delete:
                "You are about to delete the “{name}” dashboard. Are you sure? | " +
                "You are about to delete {count} dashboards. Are you sure?",
            deleted: "Dashboard successfully deleted | Dashboards successfully deleted",
            empty: "Dashboard is empty",
            loading: "Loading dashboards…",
            none: "No dashboards defined",
            notFound: "Dashboard not found",
            saved: "Dashboard successfully saved",
            selected: "{0} dashboard selected | {0} dashboards selected",
        },
        data: {
            none: "No data found",
        },
        database: {
            invalidFile: "Invalid file type",
            restore:
                "You are about to restore the database from an archive. All existing data <u>will be lost</u>. " +
                "Are you sure?",
            restored: "Database successfully restored",
            restoreFailed: "Cannot restore database: {0}",
        },
        documentation:
            "Documentation regarding Facette’s architecture, its usage and REST API is available on a dedicated " +
            "website.",
        error: {
            _: "Error: {0}",
            bulk: "An error occurred during bulk execution",
            formVerify: "Please verify provided information",
            notFound: "Resource not found",
            unhandled: "An unhandled error has occurred",
        },
        filters: {
            none: "No provider filters defined",
        },
        keyboard: {
            shortcutsDisabled: "Keyboard shortcuts are currently disabled",
        },
        labels: {
            emptyDiscarded: "Labels having empty values will be discarded",
            none: "No labels found",
        },
        lastModified: "Last modified on {0}",
        markers: {
            none: "No markers defined",
        },
        metrics: {
            none: "No metrics found",
            selected: "{0} metric selected | {0} metrics selected",
        },
        notAvailable: "Not available",
        notDefined: "Not defined",
        providers: {
            conflict: "A provider with the same name already exists.",
            delete:
                "You are about to delete the “{name}” provider. Are you sure? | " +
                "You are about to delete {count} providers. Are you sure?",
            deleted: "Provider successfully deleted | Providers successfully deleted",
            disable:
                "You are about to disable the “{name}” provider. Are you sure? | " +
                "You are about to disable {count} providers. Are you sure?",
            disabled: "Provider successfully disabled | Providers successfully disabled",
            enable:
                "You are about to enable the “{name}” provider. Are you sure? | " +
                "You are about to enable {count} providers. Are you sure?",
            enabled: "Provider successfully enabled | Providers successfully enabled",
            none: "No providers defined",
            saved: "Provider successfully saved",
            selected: "{0} provider selected | {0} providers selected",
            supportFailed: "Cannot load provider support: {0}",
            test: {
                error: "Cannot validate provider: {0}",
                success: "Provider successfully validated",
            },
        },
        series: {
            emptyAxis: "No series have been associated with this axis yet",
            none: "No series defined",
        },
        settings: {
            reload: "Saving settings will trigger a page reload to apply changes",
            saved: "Settings successfully saved",
        },
        templates: {
            none: "No templates defined",
        },
        unsavedLost: "All unsaved data will be lost. Are you sure?",
        variables: {
            none: "No variables defined",
        },
    },
};
