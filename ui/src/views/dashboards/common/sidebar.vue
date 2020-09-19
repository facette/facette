<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-sidebar :active="sidebar">
        <v-toolbar clip="sidebar">
            <v-select
                :options="sections"
                :searchable="false"
                @update:value="onSection"
                v-model:value="section"
                v-if="$route.name === 'basket-show' || $route.name === 'dashboards-home'"
            ></v-select>

            <v-button
                icon="arrow-left"
                :to="{name: 'dashboards-home'}"
                v-shortcut="{keys: 'alt+up', help: i18n.t('labels.goto.home'), tooltipHelp: false}"
                v-else
            >
                {{ i18n.t("labels.goto.home") }}
            </v-button>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <template v-else>
            <v-label>{{ title }}</v-label>

            <v-message type="placeholder" v-if="dashboardsLoading">
                <v-spinner :size="16"></v-spinner>
                {{ i18n.t("messages.dashboards.loading") }}
            </v-message>

            <v-button
                icon="folder"
                :key="index"
                :to="{name: 'dashboards-show', params: {id: dashboard.name}}"
                v-for="(dashboard, index) in dashboards"
                v-else-if="$route.name.startsWith('dashboards-')"
            >
                {{ dashboardLabel(dashboard) }}
            </v-button>

            <v-button disabled icon="info-circle" v-if="$route.name !== 'dashboards-home' && !dashboard?.items?.length">
                {{ i18n.t(`messages.${type}.${error === "notFound" ? "notFound" : "empty"}`) }}
            </v-button>

            <v-button
                :class="{active: index === highlightIndex}"
                :href="`#item${index}`"
                :key="index"
                @click="highlight($event, index)"
                v-for="(item, index) in dashboard?.items"
                v-else-if="dashboard"
            >
                {{ itemLabel(item) }}
            </v-button>
        </template>
    </v-sidebar>
</template>

<script lang="ts">
import {computed, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useRouter} from "vue-router";
import {useStore} from "vuex";

import {SelectOption} from "types/ui";

import common from "@/common";
import {useUI} from "@/components/ui";
import api from "@/lib/api";
import {dataFromVariables} from "@/lib/objects";
import {parseVariables, renderTemplate} from "@/lib/template";
import {State} from "@/store";

const defaultSection = "home";

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();
        const ui = useUI();

        const {erred, error, loading, onFetchRejected, sidebar} = common;

        const dashboards = ref<Array<Dashboard>>([]);
        const dashboardsLoading = ref(false);
        const section = ref(defaultSection);

        const sections = ref<Array<SelectOption>>([
            {label: i18n.t("labels.home"), value: "home", icon: "home"},
            {label: i18n.t("labels.basket._"), value: "basket", icon: "shopping-basket"},
        ]);

        const dashboard = computed(() => {
            return (store.state.routeData?.dashboard ?? null) as Dashboard | null;
        });

        const dashboardRefs = computed(() => {
            return (store.state.routeData?.dashboardRefs ?? {}) as Record<string, unknown>;
        });

        const highlightIndex = computed(() => {
            return (store.state.routeData?.highlightIndex ?? null) as number | null;
        });

        const type = computed(() => {
            return (store.state.routeData?.type ?? null) as string | null;
        });

        const title = computed(() => {
            if (router.currentRoute.value.name === "basket-show") {
                return i18n.t("labels.basket._");
            } else if (dashboard.value) {
                return dashboard.value?.options?.title ?? dashboard.value.name;
            } else if (
                router.currentRoute.value.name === "dashboards-show" ||
                router.currentRoute.value.name === "charts-show"
            ) {
                return router.currentRoute.value.params.id as string;
            }

            return i18n.t("labels.dashboards._", 2);
        });

        const dashboardLabel = (dashboard: Dashboard): string => {
            if (dashboard.options?.title) {
                let label = dashboard.options.title;

                if (parseVariables(label).length > 0) {
                    label = renderTemplate(label, dataFromVariables(dashboard.options?.variables ?? []));
                }

                return label;
            }

            return dashboard.name;
        };

        const getDashboards = (): void => {
            dashboardsLoading.value = true;

            api.objects<Dashboard>("dashboards", {kind: "plain", parent: dashboard.value?.id})
                .then(response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get objects");
                    }

                    dashboards.value = response.data;
                }, onFetchRejected)
                .finally(() => {
                    dashboardsLoading.value = false;
                });
        };

        const highlight = (ev: MouseEvent, index: number | null = null): void => {
            if (ev.ctrlKey || ev.metaKey) {
                return;
            }

            if (location.hash === `#item${index}`) {
                location.hash = "";
                ev.preventDefault();
            }
        };

        const itemLabel = (item: DashboardItem): string => {
            switch (item.type) {
                case "chart": {
                    if (item.options) {
                        const ref = dashboardRefs.value[`chart|${item.options.id}`] as Chart;
                        return ref ? ref.options?.title ?? ref.name : (item.options.id as string);
                    }

                    break;
                }

                case "text": {
                    return (item.options as {content: string; title: string}).title;
                }

                default:
                    return i18n.t("labels.items.unsupported");
            }

            return i18n.t("labels.unnamed");
        };

        const onSection = (to: string) => {
            switch (to) {
                case "basket":
                    router.push({name: "basket-show"});
                    break;

                case "home":
                    router.push({name: "dashboards-home"});
                    break;
            }
        };

        watch(
            () => router.currentRoute.value.name,
            to => {
                dashboards.value = [];
                section.value = to === "basket-show" ? "basket" : "home";
            },
            {immediate: true},
        );

        watch(loading, to => {
            if (!to && !erred.value && router.currentRoute.value.name !== "basket-show") {
                getDashboards();
            }
        });

        watch(dashboard, to => ui.title(to?.options?.title ?? to?.name));

        return {
            dashboard,
            dashboardLabel,
            dashboards,
            dashboardsLoading,
            error,
            highlight,
            highlightIndex,
            i18n,
            itemLabel,
            loading,
            onSection,
            section,
            sections,
            sidebar,
            title,
            type,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";

.v-sidebar {
    @include sidebar;
}
</style>
