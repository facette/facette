<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<!-- eslint-disable vue/no-mutating-props -->

<template>
    <v-label>{{ i18n.t("labels.url") }}</v-label>
    <v-input
        required
        type="url"
        :help="i18n.t('help.providers.url', ['KairosDB'])"
        :placeholder="i18n.t('labels.placeholders.example', ['http://localhost:8080/'])"
        v-model:value="settings.url"
    ></v-input>

    <v-checkbox type="toggle" v-model:value="settings.skipVerify" v-if="secured">
        {{ i18n.t("labels.tls.skipVerify") }}
    </v-checkbox>
</template>

<script lang="ts">
import {computed} from "vue";
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
    setup(props: Record<string, any>): Record<string, unknown> {
        const i18n = useI18n();

        const secured = computed(() => {
            return Boolean(((props.settings.url ?? "") as string).startsWith("https://"));
        });

        return {
            i18n,
            secured,
        };
    },
};
</script>
