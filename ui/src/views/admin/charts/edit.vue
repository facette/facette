<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content>
        <teleport to="body">
            <v-modal-chart-marker></v-modal-chart-marker>
            <v-modal-chart-series></v-modal-chart-series>
            <v-modal-template-variable></v-modal-template-variable>
        </teleport>

        <v-toolbar clip="content">
            <v-button
                icon="save"
                :disabled="erred || loading || saving"
                @click="saveChart(true)"
                v-if="prevRoute.name === 'charts-show' && !template && modifiers.alt"
            >
                {{ i18n.t("labels.saveAndGo") }}
            </v-button>

            <v-button icon="save" :disabled="erred || loading || saving" @click="saveChart(false)" v-else>
                {{ i18n.t(`labels.${template ? "templates" : "charts"}.save`) }}
            </v-button>

            <v-button icon="trash" @click="deleteChart()" v-if="!erred && edit && modifiers.alt">
                {{ i18n.t("labels.delete") }}
            </v-button>

            <v-button :disabled="erred" @click="redirectPrev()" v-else>
                {{ i18n.t("labels.cancel") }}
            </v-button>

            <v-divider vertical></v-divider>

            <v-button icon="undo" :disabled="erred || loading || !routeGuarded" @click="reset()">
                {{ i18n.t("labels.reset") }}
            </v-button>

            <v-divider vertical></v-divider>

            <template v-if="chart && edit">
                <v-spacer></v-spacer>

                <v-label class="note" v-if="chart.modifiedAt">
                    {{ i18n.t("messages.lastModified", [formatDate(chart.modifiedAt, i18n.t("date.long"))]) }}
                </v-label>
            </template>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <v-message-error v-else-if="erred"></v-message-error>

        <template v-else-if="chart">
            <div class="preview">
                <v-form v-if="dynamicVariables.length > 0">
                    <v-select
                        :key="index"
                        :label="variable.name"
                        :options="dynamicOptions[variable.name] || []"
                        v-model:value="data[variable.name]"
                        v-for="(variable, index) in dynamicVariables"
                    >
                    </v-select>
                </v-form>

                <v-chart tooltip v-model:value="resolvedChart"></v-chart>
            </div>

            <h1 v-if="!$route.params.section">{{ i18n.t("labels.general") }}</h1>

            <v-form ref="form" class="third" v-show="!$route.params.section">
                <v-label>{{ i18n.t("labels.name._") }}</v-label>
                <v-input
                    required
                    :custom-validity="objectNameValidity('charts', chart.id)"
                    :delay="350"
                    :help="i18n.t('help.charts.name')"
                    :pattern="namePattern"
                    :placeholder="i18n.t('labels.name.choose')"
                    v-autofocus.select
                    v-model:value="chart.name"
                ></v-input>

                <template v-if="link">
                    <v-label>{{ i18n.t("labels.templates._", 1) }}</v-label>
                    <v-flex class="columns">
                        <v-select
                            required
                            :options="templates"
                            :placeholder="i18n.t('labels.templates.select')"
                            v-model:value="chart.link"
                        >
                            <template v-slot:dropdown-placeholder v-if="templates.length === 0">
                                <v-label>{{ i18n.t("messages.templates.none") }}</v-label>
                            </template>
                        </v-select>

                        <v-button
                            icon="pencil-alt"
                            :to="{name: 'admin-charts-edit', params: {id: String(chart.link)}}"
                            :style="{visibility: chart.link ? 'visible' : 'hidden'}"
                        >
                            {{ i18n.t("labels.templates.edit") }}
                        </v-button>
                    </v-flex>
                </template>

                <template v-else>
                    <v-label>{{ i18n.t("labels.title") }}</v-label>
                    <v-input
                        :delay="350"
                        :help="i18n.t('help.charts.title')"
                        v-model:value="chart.options.title"
                    ></v-input>

                    <v-label>{{ i18n.t("labels.charts.type._") }}</v-label>
                    <v-select
                        class="half"
                        required
                        :options="types"
                        :placeholder="i18n.t('labels.charts.type.select')"
                        v-model:value="chart.options.type"
                    ></v-select>

                    <v-label>{{ i18n.t("labels.charts.legend._") }}</v-label>
                    <v-checkbox type="toggle" v-model:value="chart.options.legend">
                        {{ i18n.t("labels.charts.legend.show") }}
                    </v-checkbox>
                </template>
            </v-form>

            <template v-if="$route.params.section === 'variables' && variables?.length">
                <h1>{{ i18n.t("labels.variables._") }}</h1>

                <v-form-template-variables
                    :parsed="variables"
                    v-model:value="chart.options.variables"
                ></v-form-template-variables>
            </template>

            <template v-else>
                <template v-if="$route.params.section === 'series'">
                    <h1>{{ i18n.t("labels.series._", 2) }}</h1>

                    <v-form>
                        <v-message type="info" v-if="chart.series.length === 0">
                            {{ i18n.t("messages.series.none") }}
                        </v-message>

                        <v-table class="fixed" draggable v-model:value="chart.series" v-else>
                            <template v-slot:header>
                                <v-table-cell></v-table-cell>
                                <v-table-cell grow>{{ i18n.t("labels.series._", 2) }}</v-table-cell>
                                <v-table-cell>{{ i18n.t("labels.charts.axes._", 1) }}</v-table-cell>
                                <v-table-cell></v-table-cell>
                            </template>

                            <template v-slot="series">
                                <v-table-cell class="v-table-color">
                                    <span class="color" :style="{backgroundColor: colors[series.index]}"></span>
                                </v-table-cell>

                                <v-table-cell grow>
                                    <v-highlight :content="formatExpr(series.value.expr, true)"></v-highlight>
                                </v-table-cell>

                                <v-table-cell>
                                    {{ i18n.t(`labels.charts.axes.${series.value.options?.axis || "left"}`) }}
                                </v-table-cell>

                                <v-table-cell>
                                    <v-button
                                        class="reveal"
                                        icon="pencil-alt"
                                        @click="editSeries(series.index)"
                                        v-tooltip="i18n.t('labels.series.edit')"
                                    ></v-button>

                                    <v-button
                                        class="reveal"
                                        icon="times"
                                        @click="removeSeries(series.index)"
                                        v-tooltip="i18n.t('labels.series.remove')"
                                    ></v-button>
                                </v-table-cell>
                            </template>
                        </v-table>

                        <v-toolbar>
                            <v-button icon="plus" @click="addSeries">{{ i18n.t("labels.series.add") }}</v-button>
                        </v-toolbar>
                    </v-form>
                </template>

                <template v-else-if="$route.params.section === 'axes'">
                    <h1>{{ i18n.t("labels.charts.axes._", 2) }}</h1>

                    <v-form>
                        <v-flex class="columns">
                            <v-flex direction="column">
                                <h2>{{ i18n.t("labels.charts.axes.yLeft") }}</h2>

                                <v-form-yaxis :axis="chart.options.axes.y.left"></v-form-yaxis>

                                <v-message type="warning" v-if="chart.options.axes.y.left.show && !seriesAxes.left">
                                    {{ i18n.t("messages.series.emptyAxis") }}
                                </v-message>
                            </v-flex>

                            <v-flex direction="column">
                                <h2>{{ i18n.t("labels.charts.axes.yRight") }}</h2>

                                <v-form-yaxis :axis="chart.options.axes.y.right"></v-form-yaxis>

                                <v-message type="warning" v-if="chart.options.axes.y.right.show && !seriesAxes.right">
                                    {{ i18n.t("messages.series.emptyAxis") }}
                                </v-message>
                            </v-flex>

                            <v-flex direction="column">
                                <h2>{{ i18n.t("labels.charts.axes.x") }}</h2>

                                <v-checkbox type="toggle" v-model:value="chart.options.axes.x.show">
                                    {{ i18n.t("labels.show") }}
                                </v-checkbox>
                            </v-flex>
                        </v-flex>
                    </v-form>
                </template>

                <template v-if="$route.params.section === 'markers'">
                    <h1>{{ i18n.t("labels.markers._", 2) }}</h1>

                    <v-form>
                        <v-message type="info" v-if="chart.options.markers.length === 0">
                            {{ i18n.t("messages.markers.none") }}
                        </v-message>

                        <v-table class="fixed" v-model:value="chart.options.markers" v-else>
                            <template v-slot:header>
                                <v-table-cell></v-table-cell>
                                <v-table-cell>{{ i18n.t("labels.value") }}</v-table-cell>
                                <v-table-cell>{{ i18n.t("labels.labels", 1) }}</v-table-cell>
                                <v-table-cell grow>{{ i18n.t("labels.charts.axes._", 1) }}</v-table-cell>
                                <v-table-cell></v-table-cell>
                            </template>

                            <template v-slot="marker">
                                <v-table-cell class="v-table-color">
                                    <span
                                        class="color"
                                        :style="{backgroundColor: marker.value.color || 'var(--color)'}"
                                    ></span>
                                </v-table-cell>

                                <v-table-cell>
                                    {{ marker.value.value }}
                                </v-table-cell>

                                <v-table-cell>
                                    {{ marker.value.label || marker.value.value }}
                                </v-table-cell>

                                <v-table-cell grow>
                                    <v-flex>
                                        {{ i18n.t(`labels.charts.axes.${marker.value.axis}`) }}

                                        <v-icon
                                            icon="info-circle"
                                            v-tooltip="i18n.t('messages.series.emptyAxis')"
                                            v-if="marker.value.axis && !seriesAxes[marker.value.axis]"
                                        ></v-icon>
                                    </v-flex>
                                </v-table-cell>

                                <v-table-cell>
                                    <v-button
                                        class="reveal"
                                        icon="pencil-alt"
                                        @click="editMarker(marker.index)"
                                        v-tooltip="i18n.t('labels.markers.edit')"
                                    ></v-button>

                                    <v-button
                                        class="reveal"
                                        icon="times"
                                        @click="removeMarker(marker.index)"
                                        v-tooltip="i18n.t('labels.markers.remove')"
                                    ></v-button>
                                </v-table-cell>
                            </template>
                        </v-table>

                        <v-toolbar>
                            <v-button icon="plus" @click="addMarker">{{ i18n.t("labels.markers.add") }}</v-button>
                        </v-toolbar>
                    </v-form>
                </template>
            </template>
        </template>
    </v-content>
</template>

<script lang="ts">
import {colors as boulaColors} from "@facette/boula/src/config";
import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import {computed, onBeforeMount, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {onBeforeRouteLeave, onBeforeRouteUpdate, useRouter} from "vue-router";
import {useStore} from "vuex";

import {APIResponse, Chart, ChartSeries, ChartXAxis, ChartYAxis, Marker, TemplateVariable} from "types/api";
import {FormComponent, SelectOption} from "types/ui";

import common, {namePattern} from "@/common";
import {ModalConfirmParams} from "@/components/modal/confirm.vue";
import {useUI} from "@/components/ui";
import {formatDate} from "@/helpers/date";
import {objectNameValidity} from "@/helpers/validity";
import api from "@/lib/api";
import {formatExpr} from "@/lib/expr";
import {parseChartVariables, renderChart, resolveVariables} from "@/lib/objects";
import {State} from "@/store";

import FormTemplateVariablesComponent from "../common/form/template-variables.vue";
import ModalTemplateVariableComponent from "../common/modal/template-variable.vue";
import FormYaxisComponent from "./form/yaxis.vue";
import ModalChartMarkerComponent, {ModalChartMarkerParams} from "./modal/marker.vue";
import ModalChartSeriesComponent, {ModalChartSeriesParams} from "./modal/series.vue";

const chartTypes = ["area", "bar", "line"];

const defaultXAxis: ChartXAxis = {
    show: true,
};

const defaultYAxis: ChartYAxis = {
    show: true,
    label: undefined,
    max: undefined,
    min: undefined,
    unit: {
        type: undefined,
        base: undefined,
    },
};

const defaultChart: Chart = {
    id: "",
    name: "",
    options: {
        axes: {
            x: cloneDeep(defaultXAxis),
            y: {
                center: false,
                left: cloneDeep(defaultYAxis),
                right: merge({}, defaultYAxis, {show: false}),
            },
        },
        legend: false,
        markers: [],
        title: "",
        type: "area",
        variables: [],
    },
    series: [],
};

const defaultChartLinked: Chart = {
    id: "",
    name: "",
    options: {
        variables: [],
    },
};

export default {
    components: {
        "v-form-yaxis": FormYaxisComponent,
        "v-form-template-variables": FormTemplateVariablesComponent,
        "v-modal-chart-marker": ModalChartMarkerComponent,
        "v-modal-chart-series": ModalChartSeriesComponent,
        "v-modal-template-variable": ModalTemplateVariableComponent,
    },
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();
        const ui = useUI();

        const {
            applyRouteParams,
            beforeRoute,
            erred,
            loading,
            modifiers,
            onFetchRejected,
            prevRoute,
            routeGuarded,
            watchGuard,
            unwatchGuard,
        } = common;

        const chart = ref<Chart | null>(null);
        const data = ref<Record<string, string>>({});
        const dynamicData = ref<Record<string, Array<string>>>({});
        const form = ref<FormComponent | null>(null);
        const invalid = ref(false);
        const linked = ref<Chart | null>(null);
        const saving = ref(false);
        const templates = ref<Array<SelectOption>>([]);
        const variables = ref<Array<TemplateVariable>>([]);

        const colors = computed(
            (): Array<string> =>
                chart.value?.series?.map(
                    (series, index) => series.options?.color || boulaColors[index % boulaColors.length],
                ) ?? [],
        );

        const dynamicOptions = computed(
            (): Record<string, Array<SelectOption>> => {
                return Object.keys(dynamicData.value).reduce(
                    (out: Record<string, Array<SelectOption>>, name: string) => {
                        out[name] = dynamicData.value[name].map(value => ({label: value, value}));
                        return out;
                    },
                    {},
                );
            },
        );

        const dynamicVariables = computed(
            (): Array<TemplateVariable> => {
                return chart.value?.options?.variables?.filter(variable => variable.dynamic) ?? [];
            },
        );

        const edit = computed(
            () => router.currentRoute.value.params.id !== "new" && router.currentRoute.value.params.id !== "link",
        );

        const link = computed(() => router.currentRoute.value.params.id === "link" || Boolean(chart.value?.link));

        const resolvedChart = computed((): Chart | null => {
            if (linked.value) {
                return renderChart(linked.value, data.value);
            } else if (!link.value) {
                return chart.value;
            }

            return null;
        });

        const seriesAxes = computed((): {left: boolean; right: boolean} => {
            const out = {left: false, right: false};
            chart.value?.series?.forEach(series => (out[series.options?.axis ?? "left"] = true));
            return out;
        });

        const template = computed(() => !link.value && variables.value.length > 0);

        const types = computed<Array<SelectOption>>(() =>
            chartTypes.map(type => ({label: i18n.t(`labels.charts.type.${type}`), value: type})),
        );

        const addMarker = async (): Promise<void> => {
            const marker = await ui.modal<Marker | false>("chart-marker", {
                marker: {},
            } as ModalChartMarkerParams);

            if (marker) {
                chart.value?.options?.markers?.push(marker);
            }
        };

        const addSeries = async (): Promise<void> => {
            const series = await ui.modal<ChartSeries | false>("chart-series", {
                series: {},
            } as ModalChartSeriesParams);

            if (series) {
                chart.value?.series?.push(series);
            }
        };

        const deleteChart = async (): Promise<void> => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            const ok = await ui.modal<boolean>("confirm", {
                button: {
                    label: i18n.t(`labels.charts.delete`, 1),
                    danger: true,
                },
                message: i18n.t(`messages.charts.delete`, chart.value, 1),
            } as ModalConfirmParams);

            if (ok) {
                api.delete("charts", chart.value.id).then(() => {
                    ui.notify(i18n.t(`messages.charts.deleted`, 1), "success");
                    unwatchGuard();
                    router.push({name: "admin-charts-list"});
                }, onFetchRejected);
            }
        };

        const editMarker = async (index: number): Promise<void> => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            const marker = await ui.modal<Marker>("chart-marker", {
                edit: true,
                marker: chart.value.options?.markers?.[index],
            } as ModalChartMarkerParams);

            if (marker) {
                chart.value.options?.markers?.splice(index, 1, marker);
            }
        };

        const editSeries = async (index: number): Promise<void> => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            const series = await ui.modal<ChartSeries>("chart-series", {
                edit: true,
                series: chart.value.series?.[index],
            } as ModalChartSeriesParams);

            if (series) {
                chart.value.series?.splice(index, 1, series);
            }
        };

        const redirect = (go: boolean): void => {
            router.push(
                go || prevRoute.value?.name === "charts-show"
                    ? {name: "charts-show", params: {id: chart.value?.name as string}}
                    : {name: "admin-charts-list", query: template.value ? {kind: "template"} : {}},
            );
        };

        const redirectPrev = (): void => {
            router.push(
                prevRoute.value ?? {name: "admin-charts-list", query: template.value ? {kind: "template"} : {}},
            );
        };

        const removeMarker = (index: number): void => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            chart.value?.options?.markers?.splice(index, 1);
        };

        const removeSeries = (index: number): void => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            chart.value?.series?.splice(index, 1);
        };

        const reset = async (force = false): Promise<void> => {
            if (!force) {
                const ok = await ui.modal<boolean>("confirm", {
                    button: {
                        label: i18n.t("labels.charts.reset"),
                        danger: true,
                    },
                    message: i18n.t("messages.unsavedLost"),
                } as ModalConfirmParams);

                if (!ok) {
                    return;
                }
            }

            let promise: Promise<APIResponse<Chart | null>>;
            if (edit.value) {
                promise = api.object<Chart>("charts", router.currentRoute.value.params.id as string);
            } else {
                const data = {} as Chart;

                if (link.value && router.currentRoute.value.query.template) {
                    data.link = router.currentRoute.value.query.template as string;
                } else {
                    const series = store.state.routeData?.prefill as Array<ChartSeries> | undefined;
                    if (series !== undefined) {
                        data.series = series;
                    }
                }

                promise = Promise.resolve({data: Object.keys(data).length > 0 ? data : null});
            }

            form.value?.reset();
            invalid.value = false;
            store.commit("loading", true);

            promise
                .then(async response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get chart");
                    }

                    if (router.currentRoute.value.params.id === "link" || response.data?.link) {
                        await api
                            .objects<Chart>("charts", {kind: "template"})
                            .then(response => {
                                if (response.data) {
                                    templates.value = response.data.map(chart => ({
                                        label: chart.name,
                                        value: chart.id,
                                    }));
                                }
                            });

                        chart.value = merge({}, defaultChartLinked, response.data);
                    } else {
                        chart.value = merge({}, defaultChart, response.data);
                    }

                    watchGuard(chart);
                }, onFetchRejected)
                .finally(() => {
                    updateRouteData();
                    store.commit("loading", false);
                });
        };

        const saveChart = (go: boolean): void => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            if (!routeGuarded) {
                redirect(go);
                return;
            } else if (!form.value?.checkValidity()) {
                ui.notify(i18n.t("messages.error.formVerify"), "error");
                invalid.value = true;
                updateRouteData();
                return;
            }

            const obj: Chart = cloneDeep(chart.value);
            obj.template = template.value;

            api.saveObject<Chart>("charts", obj).then(() => {
                ui.notify(i18n.t("messages.charts.saved"), "success");
                unwatchGuard();
                redirect(go);
            }, onFetchRejected);
        };

        const updateRouteData = (clear = false): void => {
            store.commit(
                "routeData",
                !clear
                    ? {
                          invalid: {
                              general: invalid.value,
                          },
                          link: link.value,
                          markers: chart.value?.options?.markers?.length ?? 0,
                          series: chart.value?.series?.length ?? 0,
                          variables: variables.value?.length ?? 0,
                      }
                    : null,
            );
        };

        onBeforeRouteLeave(beforeRoute);

        onBeforeRouteUpdate(beforeRoute);

        onBeforeMount(() => applyRouteParams());

        onMounted(() => {
            ui.title(`${i18n.t("labels.charts._", 2)} – ${i18n.t("labels.adminPanel")}`);

            reset(true);
        });

        onBeforeUnmount(() => updateRouteData(true));

        watch(
            chart,
            async to => {
                if (!to) {
                    return;
                }

                if (to.link) {
                    // Skip loading linked chart if didn't change
                    if (to.link !== linked.value?.id) {
                        await api.object<Chart>("charts", to.link).then(response => {
                            if (response.data === undefined) {
                                return Promise.reject("cannot get linked chart");
                            }

                            linked.value = response.data;
                            variables.value = parseChartVariables(response.data);
                        });
                    }
                } else {
                    variables.value = parseChartVariables(to);
                }

                dynamicData.value = await resolveVariables(to.options?.variables ?? []);

                data.value = Object.keys(dynamicData.value).reduce((out: Record<string, string>, key: string) => {
                    out[key] = dynamicData.value[key]?.[0];
                    return out;
                }, {});

                updateRouteData();
            },
            {deep: true},
        );

        watch(
            () => router.currentRoute.value.params,
            (to, from) => {
                if (to.id !== from.id) {
                    reset(true);
                } else if (to.section !== from.section) {
                    invalid.value = !form.value?.checkValidity();
                    updateRouteData();
                }
            },
        );

        return {
            addMarker,
            addSeries,
            chart,
            colors,
            data,
            deleteChart,
            dynamicOptions,
            dynamicVariables,
            edit,
            editMarker,
            editSeries,
            erred,
            form,
            formatDate,
            formatExpr,
            i18n,
            link,
            loading,
            modifiers,
            namePattern,
            objectNameValidity,
            prevRoute,
            redirectPrev,
            removeMarker,
            removeSeries,
            reset,
            resolvedChart,
            routeGuarded,
            saveChart,
            saving,
            seriesAxes,
            template,
            templates,
            types,
            variables,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";
@import "../../../components/ui/components/form/mixins";

.v-content {
    @include content;

    .preview {
        background-color: var(--background);
        border-bottom: 1px solid var(--sidebar-background);
        margin: -2.25rem -2.25rem 2.25rem;
        padding: 1rem;
        position: sticky;
        top: calc(var(--toolbar-size) * 2);
        z-index: 1;

        .v-form {
            .v-select {
                width: auto;
            }
        }

        .v-chart {
            height: 16rem;
            width: 100%;
        }
    }

    .columns {
        @include form;

        .column + .column {
            margin-left: 4rem;
        }

        .v-message {
            margin-top: 2.25rem;
        }
    }

    .color {
        border-radius: 0.1rem;
        display: inline-block;
        height: 0.6rem;
        margin-right: 0.25rem;
        min-width: 0.6rem;
        width: 0.6rem;
    }

    .v-table-color {
        padding-right: 0;
        text-align: center;
    }
}
</style>
