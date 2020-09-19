<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content>
        <v-toolbar clip="content">
            <v-spacer></v-spacer>

            <v-button href="https://facette.io" icon="globe" target="_blank">
                {{ i18n.t("labels.visit.website") }}
            </v-button>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <v-message-error @retry="getVersion" v-else-if="erred"></v-message-error>

        <template v-else>
            <h1>{{ i18n.t("labels.info._") }}</h1>

            <dl>
                <template v-if="version">
                    <dt>{{ i18n.t("labels.info.version") }}</dt>
                    <dd>{{ version.version }}</dd>

                    <dt>{{ i18n.t("labels.info.branch") }}</dt>
                    <dd>{{ version.branch }}</dd>

                    <dt>{{ i18n.t("labels.info.revision") }}</dt>
                    <dd>{{ version.revision }}</dd>

                    <dt>{{ i18n.t("labels.info.compiler") }}</dt>
                    <dd>{{ version.compiler }}</dd>

                    <dt>{{ i18n.t("labels.info.buildDate") }}</dt>
                    <dd>{{ version.buildDate }}</dd>
                </template>

                <dt>{{ i18n.t("labels.info.connectors") }}</dt>
                <template v-if="apiOptions.connectors">
                    <dd :key="name" v-for="name in apiOptions.connectors">{{ name }}</dd>
                </template>
                <dd v-else>{{ i18n.t("messages.notAvailable") }}</dd>
            </dl>
        </template>
    </v-content>
</template>

<script lang="ts">
import {computed, onMounted, ref} from "vue";
import {useI18n} from "vue-i18n";
import {useStore} from "vuex";

import {Version} from "types/api";

import common from "@/common";
import {useUI} from "@/components/ui";
import api from "@/lib/api";
import {State} from "@/store";

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const store = useStore<State>();
        const ui = useUI();

        const {erred, loading, onFetchRejected} = common;

        const version = ref<Version | null>(null);

        const apiOptions = computed(() => store.state.apiOptions);

        const getVersion = (): void => {
            store.commit("loading", true);

            api.version()
                .then(response => {
                    version.value = response.data ?? null;
                }, onFetchRejected)
                .finally(() => {
                    store.commit("loading", false);
                });
        };

        onMounted(() => {
            ui.title(i18n.t("labels.info._"));

            getVersion();
        });

        return {
            apiOptions,
            erred,
            getVersion,
            i18n,
            loading,
            version,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";

.v-content {
    @include content;
}
</style>
