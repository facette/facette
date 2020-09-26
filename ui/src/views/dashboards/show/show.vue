<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content @transitionstart.self="onTransition" @transitionend.self="onTransition">
        <teleport to="body">
            <v-modal-time-range></v-modal-time-range>
        </teleport>

        <v-toolbar clip="content">
            <v-button
                icon="sync-alt"
                :icon-badge="!erred && options.refresh ? formatValue(options.refresh, {type: 'duration'}, 0) : null"
                :disabled="loading || erred"
                v-tooltip="i18n.t(`labels.${type}.refresh`)"
            >
                <template v-slot:dropdown>
                    <template v-if="options.refresh && modifiers.alt">
                        <v-button disabled icon="">
                            {{ i18n.t("labels.refresh.next", [formatValue(countdown, {type: "duration"}, 0)]) }}
                        </v-button>

                        <v-divider></v-divider>
                    </template>

                    <v-button
                        icon="sync-alt"
                        @click="refreshDashboard"
                        v-shortcut="{keys: 'r', help: i18n.t(`labels.${type}.refresh`)}"
                    >
                        {{ i18n.t(`labels.${type}.refresh`) }}
                    </v-button>

                    <v-divider></v-divider>

                    <v-button icon="backspace" :disabled="!options.refresh" @click="setRefreshInterval(0)">
                        {{ i18n.t("labels.refresh.reset") }}
                    </v-button>

                    <v-divider></v-divider>

                    <div class="v-columns">
                        <v-button
                            :icon="value === options.refresh ? 'check' : ''"
                            :key="index"
                            @click="setRefreshInterval(value)"
                            v-for="(value, index) in intervals"
                        >
                            {{ formatValue(value, {type: "duration"}) }}
                        </v-button>
                    </div>

                    <v-divider></v-divider>

                    <v-button
                        icon="clock"
                        @click="setRefreshInterval(null)"
                        v-shortcut="{keys: 'shift+r', help: i18n.t('labels.refresh.setInterval')}"
                    >
                        {{ i18n.t("labels.custom") }}
                    </v-button>
                </template>
            </v-button>

            <v-button
                icon="calendar-alt"
                :badge="timezoneUTC ? 'UTC' : undefined"
                :class="{timerange: timeRangeSynced}"
                :disabled="loading || erred"
                v-tooltip="i18n.t('labels.timeRange.set')"
            >
                <template v-if="timeRangeSynced">
                    <span>{{ i18n.t("labels.timeRange.from") }}</span>
                    {{
                        absoluteRange
                            ? formatDate(options.timeRange.from, i18n.t("date.long"), false)
                            : options.timeRange.from
                    }}
                    <span>{{ i18n.t("labels.timeRange.to") }}</span>
                    {{
                        absoluteRange
                            ? formatDate(options.timeRange.to, i18n.t("date.long"), false)
                            : options.timeRange.to
                    }}
                </template>

                <template v-else>
                    {{ i18n.t("labels.timeRange.multiple") }}
                </template>

                <template v-slot:dropdown>
                    <v-button icon="backspace" :disabled="!canResetTimeRange" @click="resetTimeRange">
                        {{ i18n.t("labels.timeRange.reset") }}
                    </v-button>

                    <v-divider></v-divider>

                    <div class="v-columns">
                        <v-button
                            :icon="
                                range.value === options.timeRange.from && options.timeRange.to === 'now' ? 'check' : ''
                            "
                            :key="index"
                            @click="setTimeRange({from: range.value, to: 'now'})"
                            v-for="(range, index) in ranges"
                        >
                            {{ i18n.t(`labels.timeRange.units.${range.unit}`, range.amount) }}
                        </v-button>
                    </div>

                    <v-divider></v-divider>

                    <v-button
                        icon="calendar"
                        @click="setTimeRange(null)"
                        v-shortcut="{keys: 'alt+shift+r', help: i18n.t('labels.timeRange.set')}"
                    >
                        {{ i18n.t("labels.custom") }}
                    </v-button>

                    <v-divider></v-divider>

                    <v-label>{{ i18n.t("labels.options") }}</v-label>

                    <v-checkbox type="toggle" v-model:value="autoPropagate">
                        {{ i18n.t("labels.timeRange.autoPropagate") }}
                    </v-checkbox>
                </template>
            </v-button>

            <v-spacer></v-spacer>

            <template v-if="type !== 'basket' && basket.length > 0">
                <v-button dropdown-anchor="bottom-right" icon="shopping-basket" :icon-badge="basket.length">
                    <template v-slot:dropdown>
                        <v-button icon="eye" :to="{name: 'basket-show'}">
                            {{ i18n.t("labels.basket.preview") }}
                        </v-button>

                        <v-divider></v-divider>

                        <v-button icon="broom" @click="clearBasket">
                            {{ i18n.t("labels.basket.clear") }}
                        </v-button>
                    </template>
                </v-button>

                <v-divider vertical></v-divider>
            </template>

            <v-button
                class="icon"
                dropdown-anchor="bottom-right"
                icon="angle-double-down"
                v-tooltip.nowrap="i18n.t('labels.moreActions')"
                v-if="dashboard"
            >
                <template v-slot:dropdown>
                    <template v-if="$route.name === 'basket-show'">
                        <v-button
                            icon="save"
                            :disabled="basket.length === 0"
                            :to="{name: 'admin-dashboards-edit', params: {id: 'new'}, query: {from: 'basket'}}"
                            v-shortcut="{keys: 's', help: i18n.t('labels.dashboards.save')}"
                        >
                            {{ i18n.t("labels.dashboards.saveAs") }}
                        </v-button>

                        <v-divider></v-divider>

                        <v-button icon="broom" @click="clearBasket">
                            {{ i18n.t("labels.basket.clear") }}
                        </v-button>
                    </template>

                    <v-button
                        icon="pencil-alt"
                        :to="{
                            name: `admin-${type}-edit`,
                            params: {id: dashboard.id},
                            hash: !dashboard.link ? '#layout' : undefined,
                        }"
                        v-shortcut="{keys: 'e', help: i18n.t(`labels.${type}.edit`)}"
                        v-else
                    >
                        {{ i18n.t(`labels.${type}.edit`) }}
                    </v-button>
                </template>
            </v-button>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <v-message-error :type="type" @retry="getDashboard" v-else-if="erred"></v-message-error>

        <template v-else-if="dashboard">
            <v-message type="info" v-if="!dashboard.items?.length">
                {{ i18n.t(`messages.${type}.empty`) }}
            </v-message>

            <v-form v-if="dynamicVariables.length > 0">
                <v-select
                    :key="index"
                    :label="variable.name"
                    :options="dynamicOptions[variable.name] || []"
                    v-model:value="options.data[variable.name]"
                    v-for="(variable, index) in dynamicVariables"
                >
                </v-select>
            </v-form>

            <v-grid
                readonly
                ref="grid"
                :highlight-index="highlightIndex"
                @cursor-dispatch.stop="onDispatch"
                @range-dispatch.stop="onDispatch"
                v-model:layout="dashboard.layout"
                v-model:value="dashboard.items"
            >
                <template v-slot="item">
                    <v-chart
                        controls
                        tooltip
                        :dispatch-cursor="type !== 'charts'"
                        :legend="item.value.options.legend"
                        :range="options.timeRange"
                        v-model:value="resolvedRefs[`chart|${item.value.options.id}`]"
                        v-if="item.value.type === 'chart'"
                    >
                        <template v-slot:more>
                            <v-button
                                icon="minus"
                                @click="removeBasketItem(item.index)"
                                v-if="$route.name === 'basket-show'"
                            >
                                {{ i18n.t("labels.basket.remove") }}
                            </v-button>

                            <v-button icon="shopping-basket" @click="addBasketItem(item.index)" v-else>
                                {{ i18n.t("labels.basket.add") }}
                            </v-button>

                            <v-divider></v-divider>

                            <v-button
                                icon="pencil-alt"
                                :to="{
                                    name: `admin-${item.value.type}s-edit`,
                                    params: {id: item.value.options.id},
                                }"
                            >
                                {{ i18n.t(`labels.${item.value.type}s.edit`) }}
                            </v-button>
                        </template>
                    </v-chart>

                    <v-text controls v-model:value="item.value.options" v-else-if="item.value.type === 'text'">
                        <template v-slot:more>
                            <v-button
                                icon="minus"
                                @click="removeBasketItem(item.index)"
                                v-if="$route.name === 'basket-show'"
                            >
                                {{ i18n.t("labels.basket.remove") }}
                            </v-button>

                            <v-button icon="shopping-basket" @click="addBasketItem(item.index)" v-else>
                                {{ i18n.t("labels.basket.add") }}
                            </v-button>
                        </template>
                    </v-text>
                </template>
            </v-grid>
        </template>
    </v-content>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import isEqual from "lodash/isEqual";
import {DateTime} from "luxon";
import {ComponentPublicInstance, computed, onBeforeMount, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useRouter} from "vue-router";
import {useStore} from "vuex";

import {APIResponse, Chart, Dashboard, DashboardItem, Reference, TemplateVariable, TimeRange} from "types/api";
import {SelectOption} from "types/ui";

import common from "@/common";
import {dateFormatDisplay, defaultTimeRange, ranges} from "@/components/chart";
import {ModalPromptParams} from "@/components/modal/prompt.vue";
import {ModalTimeRangeParams} from "@/components/modal/time-range.vue";
import {useUI} from "@/components/ui";
import {formatDate, parseDate} from "@/helpers/date";
import {formatValue} from "@/helpers/value";
import api from "@/lib/api";
import {dataFromVariables, renderChart, resolveDashboardReferences, resolveVariables} from "@/lib/objects";
import {State} from "@/store";

interface Options {
    data: Record<string, string>;
    timeRange: TimeRange;
    refresh: number;
}

function mapReferences(refs: Array<Reference>): Record<string, unknown> {
    return (
        refs.reduce((out: Record<string, unknown>, ref: Reference) => {
            switch (ref.type) {
                case "chart":
                    out[`chart|${(ref.value as Chart).id}`] = ref.value;
                    break;
            }

            return out;
        }, {}) ?? {}
    );
}

const defaultOptions: Options = {
    data: {},
    timeRange: defaultTimeRange,
    refresh: 0,
};

const intervals: Array<number> = [
    5, // 5s
    10, // 10s
    30, // 30s
    60, // 1m
    300, // 5m
    900, // 15m
    1800, // 30m
    3600, // 1h
    10800, // 3h
    21600, // 6h
    43200, // 12h
    86400, // 1d
];

export default {
    props: {
        type: {
            required: true,
            type: String,
        },
    },
    setup(props: Record<string, any>): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();
        const ui = useUI();

        const {erred, loading, modifiers, onFetchRejected} = common;

        let refreshInterval: number | null = null;
        let refreshSuspended = false;

        const countdown = ref<number | null>(null);
        const dashboard = ref<Dashboard | null>(null);
        const dashboardRefs = ref<Record<string, unknown>>({});
        const dynamicData = ref<Record<string, Array<string>>>({});
        const grid = ref<ComponentPublicInstance | null>(null);
        const highlightIndex = ref<number | null>(null);
        const options = ref<Options>(cloneDeep(defaultOptions));
        const timeRangeSynced = ref(true);

        const absoluteRange = computed((): boolean => {
            return parseDate(options.value.timeRange.from).isValid && parseDate(options.value.timeRange.to).isValid;
        });

        const autoPropagate = computed({
            get: (): boolean => {
                return store.state.autoPropagate;
            },
            set: (value: boolean): void => {
                store.commit("autoPropagate", value);
            },
        });

        const basket = computed(() => {
            return store.state.basket;
        });

        const canResetTimeRange = computed((): boolean => {
            return !isEqual(options.value.timeRange, defaultTimeRange) || !timeRangeSynced.value;
        });

        const clearBasket = (): void => {
            store.commit("basket", []);

            if (dashboard.value !== null && router.currentRoute.value.name === "basket-show") {
                dashboard.value.items = [];
            }
        };

        const dynamicOptions = computed(
            (): Record<string, Array<SelectOption>> => {
                return Object.keys(dynamicData.value).reduce(
                    (options: Record<string, Array<SelectOption>>, name: string) => {
                        options[name] = dynamicData.value[name].map(value => ({label: value, value}));
                        return options;
                    },
                    {},
                );
            },
        );

        const dynamicVariables = computed(
            (): Array<TemplateVariable> => {
                const names: Array<string> = [];

                let variables =
                    dashboard.value?.options?.variables?.filter(variable => {
                        names.push(variable.name);
                        return variable.dynamic;
                    }) ?? [];

                Object.keys(dashboardRefs.value).forEach(key => {
                    if (key.startsWith("chart|")) {
                        const vars = (dashboardRefs.value[key] as Chart).options?.variables?.filter(variable => {
                            const keep = !names.includes(variable.name);
                            if (keep) {
                                names.push(variable.name);
                            }
                            return keep;
                        });
                        if (vars) {
                            variables = variables.concat(vars);
                        }
                    }
                });

                return variables;
            },
        );

        const resolvedRefs = computed(
            (): Record<string, unknown> => {
                const staticData = dashboard.value?.options?.variables
                    ? dataFromVariables(dashboard.value.options.variables)
                    : {};

                return Object.keys(dashboardRefs.value).reduce((refs: Record<string, unknown>, key: string) => {
                    if (key.startsWith("chart|")) {
                        refs[key] = renderChart(
                            dashboardRefs.value[key] as Chart,
                            Object.assign({}, options.value.data, staticData),
                        );
                    }
                    return refs;
                }, {});
            },
        );

        const timezoneUTC = computed(() => store.state.timezoneUTC);

        const addBasketItem = (index: number): void => {
            if (!dashboard.value?.items) {
                return;
            }

            const tmpBasket = cloneDeep(basket.value);
            const item: DashboardItem = dashboard.value.items[index];

            tmpBasket.push({
                type: item.type,
                layout: {x: 0, y: tmpBasket.length, w: 1, h: 1},
                options: Object.assign({}, item.options),
            });

            store.commit("basket", tmpBasket);
        };

        const getDashboard = (): void => {
            store.commit("loading", true);

            let promise: Promise<APIResponse<Dashboard>>;

            if (router.currentRoute.value.name === "basket-show") {
                promise = resolveDashboardReferences(basket.value).then(response => {
                    return Promise.resolve({
                        data: {
                            id: "00000000-0000-0000-0000-000000000000",
                            name: "basket",
                            options: {
                                title: i18n.t("labels.basket._"),
                            },
                            layout: {
                                columns: 1,
                                rowHeight: 260,
                                rows: basket.value.length,
                            },
                            items: basket.value,
                            references: response,
                        },
                    });
                });
            } else if (props.type === "charts") {
                promise = api
                    .resolveObject<Chart>("charts", router.currentRoute.value.params.id as string)
                    .then(response => {
                        if (response.data === undefined) {
                            return Promise.reject("cannot get chart");
                        }

                        return Promise.resolve({
                            data: {
                                id: response.data.id,
                                name: response.data.name,
                                options: {
                                    title: response.data.options?.title,
                                    variables: response.data.options?.variables,
                                },
                                layout: {
                                    columns: 1,
                                    rowHeight: 260,
                                    rows: 1,
                                },
                                items: [
                                    {
                                        type: "chart",
                                        layout: {x: 0, y: 0, w: 1, h: 1},
                                        options: {
                                            id: response.data.id,
                                        },
                                    },
                                ],
                                references: [
                                    {
                                        type: "chart",
                                        value: response.data,
                                    },
                                ],
                            },
                        });
                    });
            } else {
                promise = api.resolveObject<Dashboard>("dashboards", router.currentRoute.value.params.id as string);
            }

            promise
                .then(async response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get dashboard");
                    }

                    if (response.data.references) {
                        dashboardRefs.value = mapReferences(response.data.references);
                        delete response.data.references;
                    }

                    dashboard.value = response.data;

                    if (dynamicVariables.value.length > 0) {
                        dynamicData.value = await resolveVariables(dynamicVariables.value);

                        // Auto-select first value from dynamic data to prefill
                        // options
                        options.value.data = Object.keys(dynamicData.value).reduce(
                            (data: Record<string, string>, key: string) => {
                                data[key] = options.value.data[key] ?? dynamicData.value[key][0];
                                return data;
                            },
                            {},
                        );
                    }
                }, onFetchRejected)
                .finally(() => {
                    updateRouteData();
                    store.commit("loading", false);
                });
        };

        const onDispatch = (ev: CustomEvent): void => {
            switch (ev.type) {
                case "cursor-dispatch":
                    grid.value?.$el
                        ?.querySelectorAll(".v-chart")
                        ?.forEach((el: HTMLElement) =>
                            el.dispatchEvent(new CustomEvent("cursor-chart", {detail: ev.detail})),
                        );

                    break;

                case "range-dispatch":
                    if (!autoPropagate.value && !ev.detail.force) {
                        if (timeRangeSynced.value && !isEqual(ev.detail.range, defaultTimeRange)) {
                            timeRangeSynced.value = false;
                        }

                        return;
                    }

                    options.value.timeRange = ev.detail.range;
                    timeRangeSynced.value = true;

                    break;
            }
        };

        const onHashChange = (): void => {
            if (location.hash) {
                const idx = Number(location.hash.substr(5));
                highlightIndex.value = !isNaN(idx) ? idx : null;
            } else {
                highlightIndex.value = null;
            }

            updateRouteData();
        };

        const onTransition = (ev: TransitionEvent) => {
            grid.value?.$el.style.setProperty(
                "width",
                ev.type === "transitionstart" ? `${grid.value?.$el.clientWidth}px` : null,
            );
        };

        const onVisibilityChange = (): void => {
            refreshSuspended = document.visibilityState !== "visible";

            if (!refreshSuspended && options.value.refresh > 0) {
                refreshDashboard();
                updateRefresh();
            }
        };

        const refreshDashboard = (): void => {
            grid.value?.$el
                ?.querySelectorAll(".v-chart")
                ?.forEach((el: HTMLElement) => el.dispatchEvent(new CustomEvent("refresh-chart")));
        };

        const removeBasketItem = (index: number): void => {
            if (!dashboard.value?.layout) {
                return;
            }

            const tmpBasket = cloneDeep(basket.value);
            tmpBasket.splice(index, 1);

            // Decrement following items Y position
            for (let i: number = index; i < tmpBasket.length; i++) {
                tmpBasket[i].layout.y--;
            }

            store.commit("basket", tmpBasket);

            dashboard.value.items = basket.value;
            dashboard.value.layout.rows--;
        };

        const resetTimeRange = (): void => {
            timeRangeSynced.value = true;
            setTimeRange(cloneDeep(defaultTimeRange));
        };

        const setRefreshInterval = async (value: number | null): Promise<void> => {
            const newValue =
                value !== null
                    ? value
                    : await ui.modal<number | false>("prompt", {
                          button: {
                              label: i18n.t("labels.refresh.setInterval"),
                              primary: true,
                          },
                          input: {
                              help: i18n.t("help.refresh.interval"),
                              type: "number",
                              value: options.value.refresh,
                          },
                          message: i18n.t("labels.refresh.interval"),
                      } as ModalPromptParams);

            if (newValue !== false) {
                options.value.refresh = Number(newValue);
            }
        };

        const setTimeRange = async (range: TimeRange | null): Promise<void> => {
            const newRange =
                range !== null
                    ? range
                    : await ui.modal<TimeRange | false>("time-range", {
                          range: absoluteRange.value
                              ? {
                                    from: parseDate(options.value.timeRange.from).toFormat(dateFormatDisplay),
                                    to: parseDate(options.value.timeRange.to).toFormat(dateFormatDisplay),
                                }
                              : {
                                    from: "",
                                    to: "",
                                },
                      } as ModalTimeRangeParams);

            if (newRange !== false) {
                options.value.timeRange = newRange;
            }
        };

        const updateRefresh = (): void => {
            // Trigger/Cancel refresh
            if (refreshInterval !== null) {
                clearInterval(refreshInterval);
                refreshInterval = null;
            }

            if (options.value.refresh > 0) {
                countdown.value = options.value.refresh;

                refreshInterval = setInterval(() => {
                    if (refreshSuspended) {
                        if (refreshInterval !== null) {
                            clearInterval(refreshInterval);
                            refreshInterval = null;
                        }
                        return;
                    }

                    (countdown.value as number)--;

                    if (countdown.value === 0) {
                        refreshDashboard();
                        countdown.value = options.value.refresh;
                    }
                }, 1000);
            } else {
                countdown.value = null;
            }
        };

        const updateRouteData = (clear = false): void => {
            store.commit(
                "routeData",
                !clear
                    ? {
                          dashboard: dashboard.value,
                          dashboardRefs: resolvedRefs.value,
                          highlightIndex: highlightIndex.value,
                          type: props.type,
                      }
                    : null,
            );
        };

        onBeforeMount(() => {
            const query = router.currentRoute.value.query as Record<string, string>;

            if (query.from || query.to) {
                const from = Number(query.from);
                const to = Number(query.to);

                if (!isNaN(from) && !isNaN(to)) {
                    setTimeRange({
                        from: DateTime.fromMillis(from).toISO(),
                        to: DateTime.fromMillis(to).toISO(),
                    });
                }
            }

            if (query.refresh) {
                setRefreshInterval(Number(query.refresh) || 0);
            }

            options.value.data = Object.keys(query).reduce((data: Record<string, string>, key: string) => {
                if (key.startsWith("var:")) {
                    data[key.substr(4)] = query[key];
                }
                return data;
            }, {});

            watch(
                options,
                (to: Options): void => {
                    const query: Record<string, string> = {};

                    if (to.timeRange.from !== defaultTimeRange.from) {
                        query.from = absoluteRange.value
                            ? parseDate(to.timeRange.from).toMillis().toString()
                            : (to.timeRange.from as string);
                    }

                    if (to.timeRange.to !== defaultTimeRange.to) {
                        query.to = absoluteRange.value
                            ? parseDate(to.timeRange.to).toMillis().toString()
                            : (to.timeRange.to as string);
                    }

                    if (to.refresh > 0) {
                        query.refresh = to.refresh.toString();
                    }

                    Object.keys(to.data).forEach(label => {
                        query[`var:${label}`] = to.data[label];
                    });

                    router.replace({hash: router.currentRoute.value.hash, query});

                    updateRefresh();
                    updateRouteData();
                },
                {deep: true, immediate: true},
            );
        });

        onMounted(() => {
            window.addEventListener("hashchange", onHashChange);
            document.addEventListener("visibilitychange", onVisibilityChange);

            if (location.hash) {
                onHashChange();
            }

            getDashboard();
        });

        onBeforeUnmount(() => {
            window.removeEventListener("hashchange", onHashChange);
            document.removeEventListener("visibilitychange", onVisibilityChange);

            updateRouteData(true);
        });

        watch(
            () => router.currentRoute.value.path,
            () => {
                dashboard.value = null;
                dashboardRefs.value = {};
                store.commit("loading", true);

                if (router.currentRoute.value.name?.toString().endsWith("-show")) {
                    getDashboard();
                }
            },
        );

        return {
            absoluteRange,
            addBasketItem,
            autoPropagate,
            basket,
            canResetTimeRange,
            clearBasket,
            countdown,
            dashboard,
            dynamicOptions,
            dynamicVariables,
            erred,
            formatDate,
            formatValue,
            getDashboard,
            grid,
            highlightIndex,
            i18n,
            intervals,
            loading,
            modifiers,
            onDispatch,
            onTransition,
            options,
            ranges,
            refreshDashboard,
            removeBasketItem,
            resetTimeRange,
            resolvedRefs,
            setRefreshInterval,
            setTimeRange,
            timeRangeSynced,
            timezoneUTC,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";

.v-content {
    @include content;

    padding: 1rem;

    .v-toolbar .v-button.timerange {
        ::v-deep(.v-button-content) {
            span {
                margin: 0 0.35rem;

                &:first-child {
                    margin-left: 0;
                }

                &:nth-child(2) {
                    text-transform: lowercase;
                }
            }
        }
    }

    .v-form {
        margin-bottom: 1rem;

        .v-select {
            width: auto;
        }
    }

    ::v-deep(.v-grid-anchor) {
        display: block;
        height: 0;
        transform: translateY(calc(-2 * var(--toolbar-size) - 1rem));
        width: 0;
    }
}
</style>
