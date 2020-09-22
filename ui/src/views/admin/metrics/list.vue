<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content>
        <teleport to="body">
            <v-modal-chart-preview></v-modal-chart-preview>
        </teleport>

        <v-toolbar clip="content">
            <v-button
                icon="eye"
                :disabled="!options.expr || metrics.length === 0"
                :icon-badge="options.expr ? total : null"
                @click="previewChart"
            >
                {{ i18n.t("labels.charts.preview") }}
            </v-button>

            <v-divider vertical></v-divider>

            <v-button icon="sync-alt" @click="getMetrics" v-shortcut="{keys: 'r', help: i18n.t('labels.refresh.list')}">
                {{ i18n.t("labels.refresh._") }}
            </v-button>
        </v-toolbar>

        <h1>{{ i18n.t("labels.metrics._", 2) }}</h1>

        <div class="expr">
            <v-icon icon="list-alt"></v-icon>

            <v-highlight :content="options.expr" v-if="options.expr"></v-highlight>

            <div class="placeholder" v-else>{{ i18n.t("labels.expr.none") }}</div>
        </div>

        <div class="selector">
            <div class="explorer">
                <div class="top">
                    <v-label>{{ i18n.t("labels.labels.explorer") }}</v-label>

                    <v-input
                        icon="search"
                        type="search"
                        :delay="350"
                        :placeholder="i18n.t('labels.labels.search')"
                        v-model:value="options.filter"
                        v-shortcut="{keys: 's', help: i18n.t('labels.labels.search')}"
                    ></v-input>
                </div>

                <v-spinner :size="24" v-if="accordion === null"></v-spinner>

                <v-message class="placeholder" v-else-if="Object.keys(labels).length === 0">
                    {{ i18n.t("messages.labels.none") }}
                </v-message>

                <template v-else>
                    <template :key="name" v-for="(entry, name, index) in sortLabels(labels)">
                        <v-divider v-if="index > 0"></v-divider>

                        <v-button
                            class="label"
                            :badge="entry.total"
                            :icon="accordion[name] ? 'angle-up' : 'angle-down'"
                            @click="toggleAccordion(name)"
                        >
                            {{ name }}
                        </v-button>

                        <div class="values" v-show="accordion[name]">
                            <v-button
                                :icon="hasEqMatcherCond(name, value) ? 'check-circle' : ''"
                                :key="index"
                                @click="toggleMatcher(name, value)"
                                v-tooltip="value"
                                v-for="(value, index) in entry.values"
                            >
                                {{ value }}
                            </v-button>
                        </div>

                        <v-button class="more" @click="showMore(name)" v-if="entry.values.length < entry.total">
                            {{ i18n.t("labels.show.more") }}
                        </v-button>
                    </template>
                </template>
            </div>

            <div class="list">
                <div class="top">
                    <v-label>{{ i18n.t("labels.results") }}</v-label>
                </div>

                <v-message type="info" v-if="metrics.length === 0">{{ i18n.t("messages.metrics.none") }}</v-message>

                <v-table ref="table" v-model:value="metrics" v-else>
                    <template v-slot="metric">
                        <v-table-cell class="monospace" grow>
                            <v-highlight :content="formatExpr(metric.value, true)"></v-highlight>
                        </v-table-cell>

                        <v-table-cell>
                            <v-button
                                class="reveal icon"
                                icon="far/copy"
                                @click="clipboardCopy(metric.value.toString())"
                                v-tooltip="i18n.t('labels.clipboard.copy')"
                            ></v-button>
                        </v-table-cell>
                    </template>
                </v-table>

                <v-spinner ref="spinner" :size="24" v-if="loading || page < pages"></v-spinner>
            </div>
        </div>
    </v-content>
</template>

<script lang="ts">
import {ComponentPublicInstance, computed, onBeforeMount, onMounted, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useRouter} from "vue-router";
import {useStore} from "vuex";

import {LabelValues} from "types/api";

import api from "@/lib/api";
import {formatExpr} from "@/lib/expr";
import {Labels, Matcher, Op, matcherToString, parseMatcher} from "@/lib/labels";
import common from "@/common";
import {useUI} from "@/components/ui";
import {State} from "@/store";

import ModalChartPreviewComponent, {ModalChartPreviewParams} from "./modal/chart-preview.vue";

interface Options {
    expr: string;
    filter: string;
}

const defaultOptions: Options = {
    expr: "",
    filter: "",
};

const limit = 10;

export default {
    components: {
        "v-modal-chart-preview": ModalChartPreviewComponent,
    },
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();
        const ui = useUI();

        const {erred, loading, onFetchRejected} = common;

        let intersection: IntersectionObserver | null = null;

        const accordion = ref<Record<string, boolean> | null>(null);
        const labels = ref<Record<string, LabelValues>>({});
        const matcher = ref<Matcher>([]);
        const metrics = ref<Array<Labels>>([]);
        const options = ref(Object.assign({}, defaultOptions));
        const page = ref(1);
        const spinner = ref<ComponentPublicInstance | null>(null);
        const total = ref(0);

        const pages = computed(() => Math.ceil(total.value / limit));

        const clipboardCopy = (value: string): void => {
            navigator.clipboard.writeText(value).then(() => ui.notify(i18n.t("messages.copied"), "success"));
        };

        const getMetrics = (metricsAppend = false): void => {
            if (!metricsAppend) {
                accordion.value = null;

                api.labelValues({
                    filter: options.value.filter || undefined,
                    limit,
                    match: options.value.expr || undefined,
                }).then(response => {
                    if (response.data === undefined) {
                        return;
                    }

                    labels.value = response.data;

                    accordion.value = Object.keys(response.data).reduce((out: Record<string, boolean>, key: string) => {
                        out[key] = true;
                        return out;
                    }, {});
                });
            }

            store.commit("loading", true);

            api.metrics({
                limit,
                offset: (page.value - 1) * limit,
                match: options.value.expr || undefined,
            })
                .then(response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get metrics");
                    }

                    if (metricsAppend) {
                        metrics.value = metrics.value.concat(response.data);
                    } else {
                        metrics.value = response.data;
                        page.value = 1;
                        total.value = response.total ?? 0;
                    }
                }, onFetchRejected)
                .finally(() => {
                    store.commit("loading", false);
                });
        };

        const hasEqMatcherCond = (name: string, value: string): boolean => {
            return (
                matcher.value.findIndex(a => a.op === Op.EQ && a.name === name && a.value === JSON.stringify(value)) !==
                -1
            );
        };

        const previewChart = (): void => {
            ui.modal("chart-preview", {
                expr: options.value.expr,
            } as ModalChartPreviewParams);
        };

        const showMore = (name: string): void => {
            api.labelValues({
                limit,
                name,
                offset: labels.value[name].values.length,
            }).then(response => {
                if (response.data !== undefined) {
                    labels.value[name].values = labels.value[name].values.concat(response.data[name].values);
                }
            });
        };

        const sortLabels = (labels: Record<string, LabelValues>): Record<string, LabelValues> => {
            return Object.keys(labels)
                .sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()))
                .reduce((out: Record<string, LabelValues>, name: string) => {
                    out[name] = labels[name];
                    return out;
                }, {});
        };

        const toggleAccordion = (name: string): void => {
            if (accordion.value !== null) {
                accordion.value[name] = !accordion.value[name];
            }
        };

        const toggleMatcher = (name: string, value: string): void => {
            value = JSON.stringify(value);

            const idx = matcher.value.findIndex(m => m.name === name);
            if (idx !== -1) {
                console.debug(matcher.value[idx].value, value);
                if (matcher.value[idx].value === value) {
                    matcher.value.splice(idx, 1);
                } else {
                    matcher.value[idx].op = Op.EQ;
                    matcher.value[idx].value = value;
                }
            } else {
                matcher.value.push({op: Op.EQ, name, value});
            }

            options.value = {
                filter: "",
                expr: matcherToString(matcher.value),
            };
        };

        onBeforeMount(() => {
            const query = router.currentRoute.value.query as Record<string, string>;

            if (query.filter) {
                options.value.filter = query.filter;
            }

            if (query.expr) {
                try {
                    matcher.value = parseMatcher(query.expr);
                    options.value.expr = query.expr;
                } catch (e) {}
            }
        });

        onMounted(() => {
            ui.title(`${i18n.t("labels.metrics._", 2)} – ${i18n.t("labels.adminPanel")}`);

            watch(
                options,
                (to: Options): void => {
                    const query: Record<string, string> = {};

                    if (to.filter !== "") {
                        query.filter = to.filter;
                    }

                    if (to.expr !== "") {
                        query.expr = to.expr;
                    }

                    router.replace({query});

                    getMetrics();
                },
                {deep: true, immediate: true},
            );
        });

        watch(pages, to => {
            if (page.value < to) {
                if (intersection === null && spinner.value !== null) {
                    intersection = new IntersectionObserver(
                        entries => {
                            if (entries[0].intersectionRatio > 0) {
                                page.value++;
                                getMetrics(true);
                            }
                        },
                        {threshold: 0},
                    );

                    intersection.observe(spinner.value.$el);
                }
            } else {
                intersection?.disconnect();
                intersection = null;
            }
        });

        return {
            accordion,
            clipboardCopy,
            erred,
            formatExpr,
            getMetrics,
            hasEqMatcherCond,
            i18n,
            labels,
            limit,
            loading,
            matcher,
            metrics,
            options,
            page,
            pages,
            previewChart,
            showMore,
            sortLabels,
            spinner,
            toggleAccordion,
            toggleMatcher,
            total,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";

.v-content {
    @include content;

    max-height: calc(100vh - var(--toolbar-size) * 2);

    .expr {
        align-items: center;
        background-color: var(--input-background);
        border-radius: 0.2rem;
        display: flex;
        line-height: 2.5rem;
        margin-bottom: 1.5rem;
        padding: 0 1rem;

        .v-icon {
            color: var(--gray);
            margin-right: 1rem;
        }

        .placeholder {
            opacity: var(--placeholder-opacity);
        }
    }

    .selector {
        display: flex;

        .explorer,
        .list {
            height: calc(100vh - var(--toolbar-size) * 2 - var(--content-padding) * 2 - 7.625rem);

            .top {
                border-bottom: 1px solid var(--divider-background);
                color: var(--table-header-color);
                position: sticky;
                text-transform: uppercase;
                top: 0;
                z-index: 1;
            }
        }

        .explorer {
            background-color: var(--grid-item-background);
            border-radius: 0.2rem;
            margin-right: var(--content-padding);
            overflow-y: auto;
            padding-bottom: 0.25rem;
            position: relative;
            width: 20rem;

            .top {
                background-color: var(--grid-item-background);
                border-radius: 0.2rem 0.2rem 0 0;
                margin-bottom: 0.25rem;
                padding: 0 1rem;

                .v-input {
                    margin-bottom: 0.75rem;
                    width: 100%;
                }
            }

            .v-spinner,
            .v-message.placeholder {
                bottom: 0;
                left: 0;
                position: absolute;
                right: 0;
                top: var(--table-row-height);
            }

            .v-message.placeholder {
                justify-content: center;
                color: var(--gray);
            }

            .v-divider {
                margin: 0.25rem 0;
            }

            .v-button {
                display: block;
                margin: 0;

                ::v-deep(.v-button-content) {
                    border-radius: 0;
                    justify-content: flex-start;

                    .v-button-label {
                        display: unset !important;
                        overflow: hidden;
                        text-overflow: ellipsis;
                    }
                }

                &.more ::v-deep(.v-button-label) {
                    opacity: 0.65;
                    text-align: center;
                }

                &.label {
                    ::v-deep() {
                        .v-button-content {
                            padding-right: 1rem;
                        }

                        .v-icon {
                            opacity: 0.5;
                        }
                    }
                }
            }
        }

        .list {
            overflow-y: auto;

            .top {
                background-color: var(--background);
                padding: 0 0.75rem;
            }

            .v-message,
            .v-table {
                width: calc(100vw - var(--sidebar-width) - var(--content-padding) * 3 - 20rem);
            }

            .v-table ::v-deep() {
                tr:first-child {
                    border-top-color: transparent;
                }

                .monospace {
                    font-size: 0.8rem;
                    white-space: normal;
                }
            }

            .v-spinner {
                display: block;
                height: 1.5rem;
                margin-top: 1rem;
                text-align: center;
            }
        }
    }
}
</style>
