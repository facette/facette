<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<!-- eslint-disable vue/no-mutating-props -->

<template>
    <v-label>{{ i18n.t("labels.providers.rrdtool.path") }}</v-label>
    <v-input
        required
        :help="i18n.t('help.providers.rrdtool.path')"
        :placeholder="i18n.t('labels.placeholders.example', ['/var/lib/collectd/rrd'])"
        v-model:value="settings.path"
    ></v-input>

    <v-label>{{ i18n.t("labels.providers.rrdtool.pattern") }}</v-label>
    <v-input
        required
        :help="i18n.t('help.providers.rrdtool.pattern')"
        :placeholder="i18n.t('labels.placeholders.example', ['(?P<instance>[^/]+)/(?P<__name__>.+)\\.rrd'])"
        v-model:value="settings.pattern"
    ></v-input>

    <v-label>{{ i18n.t("labels.providers.rrdtool.daemon") }}</v-label>
    <v-input
        required
        :help="i18n.t('help.providers.rrdtool.daemon')"
        :placeholder="i18n.t('labels.placeholders.example', ['unix:/var/run/rrdcached.sock'])"
        v-model:value="settings.daemon"
    ></v-input>
</template>

<script lang="ts">
import {useI18n} from "vue-i18n";

import {Provider} from "types/api";

export default {
    inheritAttrs: false,
    props: {
        settings: {
            required: true,
            type: Object as () => Provider["connector"]["settings"],
        },
    },
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        return {
            i18n,
        };
    },
};
</script>
