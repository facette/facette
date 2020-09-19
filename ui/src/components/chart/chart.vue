<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div
        class="v-chart"
        ref="el"
        :class="{empty: !value || !series}"
        @cursor-chart="onCursorChart"
        @range-chart="onRangeChart"
        @refresh-chart="update()"
    >
        <v-spinner :size="16" v-if="loading"></v-spinner>

        <v-message v-else-if="!value || !series">{{ i18n.t("messages.data.none") }}</v-message>

        <div class="v-chart-controls" v-if="controls">
            <v-toolbar v-if="!series">
                <v-button icon="sync-alt" @click="update()" v-tooltip="i18n.t('labels.charts.refresh')"></v-button>
            </v-toolbar>

            <v-toolbar v-else>
                <v-button icon="sync-alt" @click="update()" v-tooltip="i18n.t('labels.charts.refresh')"></v-button>

                <v-button
                    dropdown-anchor="bottom-right"
                    icon="calendar-alt"
                    v-tooltip="i18n.t('labels.timeRange.set')"
                    v-if="!autoPropagate"
                >
                    <template v-slot:dropdown>
                        <v-button icon="history" :disabled="!canResetTimeRange" @click="resetTimeRange">
                            {{ i18n.t("labels.timeRange.reset") }}
                        </v-button>

                        <v-divider></v-divider>

                        <div class="v-columns">
                            <v-button
                                :icon="range.value === timeRange.from && timeRange.to === 'now' ? 'check' : ''"
                                :key="index"
                                @click="setTimeRange({from: range.value, to: 'now'})"
                                v-for="(range, index) in ranges"
                            >
                                {{ i18n.t(`labels.timeRange.units.${range.unit}`, range.amount) }}
                            </v-button>
                        </div>

                        <v-divider></v-divider>

                        <v-button icon="calendar" @click="setTimeRange(null)">
                            {{ i18n.t("labels.custom") }}
                        </v-button>
                    </template>
                </v-button>

                <v-divider vertical></v-divider>

                <v-button
                    icon="search-minus"
                    @click="updateRange('zoom-out')"
                    v-tooltip="i18n.t('labels.charts.zoom.out')"
                ></v-button>

                <v-button
                    icon="search-plus"
                    @click="updateRange('zoom-in')"
                    v-tooltip="i18n.t('labels.charts.zoom.in')"
                ></v-button>

                <v-button
                    icon="arrows-alt-h"
                    @click="updateRange('propagate')"
                    v-tooltip="i18n.t('labels.timeRange.propagate')"
                    v-if="!autoPropagate"
                ></v-button>

                <v-divider vertical></v-divider>

                <v-button
                    class="icon"
                    dropdown-anchor="bottom-right"
                    icon="angle-double-down"
                    v-tooltip.nowrap="i18n.t('labels.moreActions')"
                >
                    <template v-slot:dropdown>
                        <v-button dropdown-anchor="right" icon="file-download">
                            {{ i18n.t("labels.export._") }}
                            <template v-slot:dropdown>
                                <v-button @click="downloadExport('png')">
                                    {{ i18n.t("labels.export.imagePNG") }}
                                </v-button>

                                <v-divider></v-divider>

                                <v-button @click="downloadExport('csv')">
                                    {{ i18n.t("labels.export.summaryCSV") }}
                                </v-button>

                                <v-button @click="downloadExport('json')">
                                    {{ i18n.t("labels.export.summaryJSON") }}
                                </v-button>
                            </template>
                        </v-button>

                        <template v-if="hasMore">
                            <v-divider></v-divider>

                            <slot name="more"></slot>
                        </template>
                    </template>
                </v-button>
            </v-toolbar>

            <div class="v-chart-sliders">
                <div class="range">
                    <v-button
                        class="backward"
                        icon="arrow-left"
                        :class="{active: sliders.backward}"
                        @click="updateRange('backward')"
                    ></v-button>

                    <v-button
                        class="forward"
                        icon="arrow-right"
                        :class="{active: sliders.forward}"
                        @click="updateRange('forward')"
                    ></v-button>
                </div>
            </div>
        </div>

        <div ref="chart" v-show="value"></div>
    </div>
</template>

<script lang="ts">
import Boula, * as boula from "@facette/boula";
import {timeDay, timeHour, timeMinute, timeMonth, timeSecond, timeYear} from "d3-time";
import {timeFormat, utcFormat} from "d3-time-format";
import cloneDeep from "lodash/cloneDeep";
import isEqual from "lodash/isEqual";
import {DateTime} from "luxon";
import ResizeObserver from "resize-observer-polyfill";
import slugify from "slugify";
import {SetupContext, computed, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useStore} from "vuex";

import {Chart, Series, SeriesQuery, SeriesResult, SeriesSummary, TimeRange} from "types/api";

import common from "@/common";
import {ModalTimeRangeParams} from "@/components/modal/time-range.vue";
import {useUI} from "@/components/ui";
import {parseDate} from "@/helpers/date";
import {formatValue} from "@/helpers/value";
import api from "@/lib/api";
import {State} from "@/store";

import {dateFormatDisplay, dateFormatFilename, defaultTimeRange, ranges} from ".";

interface CursorDispatchEvent {
    date: Date | null;
    el: HTMLElement;
}

interface RangeDispatchEvent {
    force?: boolean;
    range: TimeRange;
}

const mouseRange = 40;

export default {
    props: {
        controls: {
            default: false,
            type: Boolean,
        },
        dispatchCursor: {
            default: false,
            type: Boolean,
        },
        range: {
            default: null,
            type: Object as () => TimeRange,
        },
        tooltip: {
            default: false,
            type: Boolean,
        },
        value: {
            required: true,
            validator: (prop: unknown): boolean => typeof prop === "object" || prop === null,
        },
    },
    setup(props: Record<string, any>, ctx: SetupContext): Record<string, unknown> {
        const i18n = useI18n();
        const store = useStore<State>();
        const ui = useUI();

        const {onFetchRejected} = common;

        let domRect: DOMRect | null = null;
        let instance: boula.Chart | null = null;
        let intersection: IntersectionObserver | null = null;
        let lastDate: Date | null = null;
        let resize: ResizeObserver | null = null;

        const chart = ref<HTMLElement | null>(null);
        const data = ref<SeriesResult | null>(null);
        const el = ref<HTMLElement | null>(null);
        const hasMore = ref(false);
        const loading = ref(true);
        const sliders = ref<{backward: boolean; forward: boolean}>({backward: false, forward: false});
        const timeRange = ref<TimeRange | null>(props.range);

        const absoluteRange = computed((): boolean => {
            return (
                timeRange.value !== null &&
                parseDate(timeRange.value.from).isValid &&
                parseDate(timeRange.value.to).isValid
            );
        });

        const autoPropagate = computed((): boolean => {
            return store.state.autoPropagate;
        });

        const canResetTimeRange = computed((): boolean => {
            return !isEqual(timeRange.value, defaultTimeRange);
        });

        const series = computed((): boolean => {
            return !loading.value && Boolean(data.value?.series);
        });

        const timezoneUTC = computed((): boolean => {
            return store.state.timezoneUTC;
        });

        const draw = (): void => {
            if (!chart.value) {
                throw Error("cannot get chart");
            }

            const value = props.value as Chart;

            const markers: Array<boula.Marker> =
                value.options?.markers?.map(marker => ({
                    value: marker.value,
                    label: marker.label || true,
                    color: marker.color,
                    axis: marker.axis,
                })) ?? [];

            let min = NaN;
            let max = NaN;
            data.value?.series?.forEach(series => {
                series?.points?.forEach(point => {
                    if (point[1]) {
                        min = isNaN(min) ? point[1] : Math.min(min, point[1]);
                        max = isNaN(max) ? point[1] : Math.max(max, point[1]);
                    }
                });
            });

            const axisMin = value.options?.axes?.y?.left?.min;
            if (axisMin !== undefined && min < axisMin) {
                markers.push({dashed: true, value: axisMin});
            }

            const axisMax = value.options?.axes?.y?.left?.max;
            if (axisMax !== undefined && max > axisMax) {
                markers.push({dashed: true, value: axisMax});
            }

            let series: Array<boula.Series>;
            if (data.value?.series) {
                const disabledSeries: Array<string> =
                    instance?.config.series.reduce((names: Array<string>, s: boula.Series) => {
                        if (s.disabled && s.label) {
                            names.push(s.label);
                        }
                        return names;
                    }, []) ?? [];

                series = data.value.series.map((series, index) => {
                    const bs: boula.Series = {
                        label: series.name,
                        points: series.points.map(p => ({0: p[0] * 1000, 1: p[1]})),
                    };
                    if (disabledSeries.includes(series.name)) {
                        bs.disabled = true;
                    }

                    const options = value.series?.[index]?.options;
                    if (options?.color) {
                        bs.color = options.color;
                    }
                    if (options?.axis) {
                        bs.axis = options.axis;
                    }

                    return bs;
                });
            } else {
                series = [];
            }

            const events: boula.Config["events"] = {};

            if (props.dispatchCursor) {
                events.cursor = date => {
                    if (date === lastDate) {
                        return;
                    }

                    el.value?.dispatchEvent(
                        new CustomEvent<CursorDispatchEvent>("cursor-dispatch", {
                            bubbles: true,
                            detail: {
                                date,
                                el: el.value,
                            },
                        }),
                    );

                    lastDate = date;
                };
            }

            if (props.controls) {
                events.select = (from, to) => {
                    if (to > from) {
                        el.value?.dispatchEvent(
                            new CustomEvent<RangeDispatchEvent>("range-dispatch", {
                                bubbles: true,
                                detail: {
                                    range: {
                                        from: DateTime.fromJSDate(from).toISO(),
                                        to: DateTime.fromJSDate(to).toISO(),
                                    },
                                },
                            }),
                        );
                    }
                };
            }

            const config: boula.Config = {
                axes: {
                    x: {
                        draw: value?.options?.axes?.x?.show ?? true,
                        grid: false,
                        max: (data.value && parseDate(data.value.to).valueOf()) || undefined,
                        min: (data.value && parseDate(data.value.from).valueOf()) || undefined,
                        ticks: {
                            count: Math.max(Math.floor(chart.value.clientWidth / 80), 2),
                            format: (date: Date): string => {
                                const format: (specifier: string) => (date: Date) => string = timezoneUTC.value
                                    ? utcFormat
                                    : timeFormat;

                                return (timeSecond(date) < date
                                    ? format(".%L")
                                    : timeMinute(date) < date
                                    ? format(":%S")
                                    : timeHour(date) < date
                                    ? format("%H:%M")
                                    : timeDay(date) < date
                                    ? format("%H:00")
                                    : timeMonth(date) < date
                                    ? format("%a %d")
                                    : timeYear(date) < date
                                    ? format("%B")
                                    : timeFormat("%Y"))(date);
                            },
                        },
                    },
                    y: {
                        center: value.options?.axes?.y?.center ?? false,
                        left: {
                            draw: value?.options?.axes?.y?.left?.show ?? true,
                            max: Number(value.options?.axes?.y?.left?.max) || undefined,
                            min: Number(value.options?.axes?.y?.left?.min) || undefined,
                            label: {
                                text: value.options?.axes?.y?.left?.label,
                            },
                            ticks: {
                                draw: false,
                                format: v => formatValue(v, value.options?.axes?.y?.left?.unit),
                            },
                        },
                        right: {
                            draw: value?.options?.axes?.y?.right?.show ?? true,
                            max: Number(value.options?.axes?.y?.right?.max) || undefined,
                            min: Number(value.options?.axes?.y?.right?.min) || undefined,
                            label: {
                                text: value.options?.axes?.y?.right?.label,
                            },
                            ticks: {
                                draw: false,
                                format: v => formatValue(v, value.options?.axes?.y?.right?.unit),
                            },
                        },
                        stack: value.options?.axes?.y?.stack || false,
                    },
                },
                bindTo: chart.value,
                cursor: {
                    enabled: props.tooltip,
                },
                events,
                legend: {
                    enabled: value.options?.legend,
                },
                markers,
                selection: {
                    enabled: props.controls,
                },
                series,
                tooltip: {
                    enabled: props.tooltip,
                    date: {
                        format: d => DateTime.fromJSDate(d).toFormat(i18n.t("date.long")),
                    },
                },
                title: {
                    text: value.options?.title,
                },
                type: value.options?.type || "area",
            };

            if (instance === null) {
                instance = new Boula(config);
            } else {
                instance.update(config);
            }

            requestAnimationFrame(() => {
                instance?.draw();

                if (resize === null) {
                    domRect = el.value?.getBoundingClientRect() ?? null;
                    observeResize();
                }
            });
        };

        const downloadExport = (type: "csv" | "json" | "png"): void => {
            if (data.value === null) {
                return;
            }

            const el: HTMLAnchorElement = document.createElement("a");

            const baseName = `${slugify(props.value.options?.title || props.value.name)}_${parseDate(
                data.value.from,
            ).toFormat(dateFormatFilename)}_${parseDate(data.value.to).toFormat(dateFormatFilename)}`;

            switch (type) {
                case "csv":
                case "json": {
                    let href: string;

                    if (type === "csv") {
                        const summary = data.value.series.reduce((out: string, series: Series, index: number) => {
                            const keys: Array<string> = Object.keys(series.summary);
                            if (index === 0) {
                                out += `name,${keys.join(",")}\n`;
                            }
                            return `${out}"${series.name}",${keys.map(k => series.summary[k]).join(",")}\n`;
                        }, "");

                        href = `data:text/csv,${encodeURIComponent(summary)}`;
                    } else {
                        const summaries = data.value.series.reduce(
                            (out: Record<string, SeriesSummary>, series: Series) => {
                                out[series.name] = series.summary;
                                return out;
                            },
                            {},
                        );

                        href = `data:application/json,${encodeURIComponent(JSON.stringify(summaries, null, "\t"))}`;
                    }

                    Object.assign(el, {download: `${baseName}.${type}`, href});

                    document.body.appendChild(el);
                    el.click();
                    document.body.removeChild(el);

                    break;
                }

                case "png": {
                    if (instance === null) {
                        throw Error("cannot get instance");
                    }

                    const dataURL = instance.canvas.toDataURL("image/png");

                    Object.assign(el, {
                        download: `${baseName}.png`,
                        href: dataURL.replace("image/png", "image/octet-stream"),
                    });

                    document.body.appendChild(el);
                    el.click();
                    document.body.removeChild(el);

                    URL.revokeObjectURL(dataURL);

                    break;
                }
            }
        };

        const observeIntersection = (): void => {
            if (el.value === null) {
                throw Error("cannot get element");
            }

            intersection = new IntersectionObserver(
                entries => {
                    if (entries[0].intersectionRatio > 0) {
                        // Stop observing and update chart using time range from
                        // store to prevent charts timeline to drift
                        unobserveIntersection();
                        update(store.state.timeRange);
                    }
                },
                {threshold: 0},
            );

            intersection.observe(el.value);
        };

        const observeResize = (): void => {
            if (chart.value === null) {
                throw Error("cannot get chart");
            }

            resize = new ResizeObserver(entries => {
                // Ensure dimensions changed before redrawing (i.e. avoid
                // drawing twice on first draw)
                if (
                    entries[0].contentRect.width !== domRect?.width ||
                    entries[0].contentRect.height !== domRect?.height
                ) {
                    domRect = el.value?.getBoundingClientRect() ?? null;
                    requestAnimationFrame(() => instance?.draw());
                }
            });

            resize.observe(chart.value);
        };

        const onCursorChart = (ev: CustomEvent<CursorDispatchEvent>): void => {
            if (!(ev.detail.el as Node).isSameNode(el.value)) {
                (instance?.components?.cursor as boula.Component & {move: (date: Date | null) => void}).move(
                    ev.detail.date,
                );
            }
        };

        const onRangeChart = (ev: CustomEvent<RangeDispatchEvent>): void => {
            timeRange.value = ev.detail.range;
            update();
        };

        const onMouse = (ev: MouseEvent): void => {
            switch (ev.type) {
                case "mouseleave": {
                    if (!(ev.relatedTarget as HTMLElement | null)?.closest(".v-chart")?.isSameNode(el.value)) {
                        sliders.value = {backward: false, forward: false};
                    }

                    break;
                }

                case "mousemove": {
                    if (domRect === null) {
                        return;
                    }

                    const x: number = ev.pageX - domRect.x;

                    if (!sliders.value.backward && !sliders.value.forward) {
                        if (x <= mouseRange) {
                            sliders.value.backward = true;
                        } else if (x >= domRect.width - mouseRange) {
                            sliders.value.forward = true;
                        }
                    } else if (x > mouseRange * 1.65 && x < domRect.width - mouseRange * 1.65) {
                        sliders.value = {backward: false, forward: false};
                    }
                }
            }
        };

        const resetTimeRange = (): void => {
            setTimeRange(cloneDeep(defaultTimeRange));
        };

        const setTimeRange = async (range: TimeRange | null): Promise<void> => {
            const newRange =
                range !== null
                    ? range
                    : await ui.modal<TimeRange | false>("time-range", {
                          range:
                              timeRange.value !== null && absoluteRange.value
                                  ? {
                                        from: parseDate(timeRange.value.from).toFormat(dateFormatDisplay),
                                        to: parseDate(timeRange.value.to).toFormat(dateFormatDisplay),
                                    }
                                  : {
                                        from: "",
                                        to: "",
                                    },
                      } as ModalTimeRangeParams);

            if (newRange !== false) {
                el.value?.dispatchEvent(
                    new CustomEvent<RangeDispatchEvent>("range-dispatch", {bubbles: true, detail: {range: newRange}}),
                );

                timeRange.value = newRange;
                update();
            }
        };

        const unobserveIntersection = (): void => {
            if (intersection !== null) {
                intersection.disconnect();
                intersection = null;
            }
        };

        const unobserveResize = (): void => {
            if (resize !== null) {
                resize.disconnect();
                resize = null;
            }
        };

        const update = (range: TimeRange | null = null): void => {
            const value = props.value as Chart | null;

            if (intersection !== null) {
                return;
            } else if (!value?.series?.length) {
                data.value = null;
                loading.value = false;
                return;
            }

            const query: SeriesQuery = Object.assign(
                {
                    exprs: value.series?.map(series => series.expr),
                },
                range ? range : timeRange.value,
            );

            loading.value = true;

            api.query(query)
                .then(response => {
                    data.value = response.data ?? null;
                    if (data.value) {
                        draw();
                    }
                }, onFetchRejected)
                .finally(() => {
                    loading.value = false;
                });
        };

        const updateRange = (mode: "backward" | "forward" | "propagate" | "zoom-in" | "zoom-out"): void => {
            if (data.value === null) {
                return;
            }

            let from = parseDate(data.value.from);
            let to = parseDate(data.value.to);
            let delta: number;

            switch (mode) {
                case "backward":
                    delta = to.diff(from, "second").seconds * 0.25;
                    from = from.minus({seconds: delta});
                    to = to.minus({seconds: delta});
                    break;

                case "forward":
                    delta = to.diff(from, "second").seconds * 0.25;
                    from = from.plus({seconds: delta});
                    to = to.plus({seconds: delta});
                    break;

                case "zoom-in":
                    delta = to.diff(from, "second").seconds * 0.25;
                    from = from.plus({seconds: delta});
                    to = to.minus({seconds: delta});
                    break;

                case "zoom-out":
                    delta = to.diff(from, "second").seconds * 0.5;
                    from = from.minus({seconds: delta});
                    to = to.plus({seconds: delta});
                    break;
            }

            const range: TimeRange = {
                from: from.toISO(),
                to: to.toISO(),
            };

            const force = mode === "propagate";

            el.value?.dispatchEvent(
                new CustomEvent<RangeDispatchEvent>("range-dispatch", {bubbles: true, detail: {force, range}}),
            );

            if (!autoPropagate.value && !force) {
                timeRange.value = range;
                update();
            }
        };

        onMounted(() => {
            hasMore.value = Boolean(ctx.slots.more);

            if (props.controls && chart.value !== null) {
                chart.value?.addEventListener("mouseleave", onMouse);
                chart.value?.addEventListener("mousemove", onMouse);
            }

            observeIntersection();
        });

        onBeforeUnmount(() => {
            if (props.controls && chart.value !== null) {
                chart.value.removeEventListener("mouseleave", onMouse);
                chart.value.removeEventListener("mousemove", onMouse);
            }

            unobserveIntersection();
            unobserveResize();

            // FIXME: find why "onBeforeUnmount" is invoked twice when chart is
            // present in v-grid component

            instance?.destroy();
            instance = null;
        });

        watch(
            () => props.range,
            to => {
                timeRange.value = to;
                update();
            },
            {deep: true},
        );

        watch(
            () => props.value,
            () => update(),
            {deep: true},
        );

        return {
            autoPropagate,
            canResetTimeRange,
            chart,
            downloadExport,
            el,
            hasMore,
            i18n,
            loading,
            onCursorChart,
            onRangeChart,
            ranges,
            resetTimeRange,
            series,
            setTimeRange,
            sliders,
            timeRange,
            update,
            updateRange,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "~@facette/boula/dist/style.css";

@import "../../views/mixins";

.v-chart {
    background-color: inherit;
    height: 100%;
    position: relative;

    &.empty {
        align-items: center;
        display: flex;
        justify-content: center;
    }

    .chart-container {
        height: 100%;
        width: 100%;
    }

    .v-spinner {
        left: 0.5rem;
        position: absolute;
        top: 0.5rem;
        z-index: 1;
    }

    .v-chart-controls {
        bottom: 0;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        left: 0;
        pointer-events: none;
        position: absolute;
        right: 0;
        top: 0;
        visibility: hidden;
        z-index: 1;

        .v-toolbar {
            @include item-toolbar;
        }

        .v-chart-sliders {
            align-items: center;
            display: flex;
            height: calc(50% + 2rem);
            flex-direction: column;
            justify-content: space-between;
            overflow: hidden;

            .v-button {
                pointer-events: auto;
                transition: transform 0.2s var(--timing-function);
                will-change: transform;

                &.active {
                    transform: none !important;
                }

                ::v-deep(.v-button-content) {
                    background-color: var(--toolbar-background);
                }
            }

            .range {
                align-items: center;
                display: flex;
                justify-content: space-between;
                width: 100%;

                .v-button {
                    height: 4rem;
                    line-height: 4rem;
                    width: 2.25rem;

                    &.backward {
                        transform: translateX(-100%);

                        ::v-deep(.v-button-content) {
                            border-radius: 0 0.2rem 0.2rem 0;
                        }
                    }

                    &.forward {
                        transform: translateX(100%);

                        ::v-deep(.v-button-content) {
                            border-radius: 0.2rem 0 0 0.2rem;
                        }
                    }

                    ::v-deep(.v-icon) {
                        font-size: 1.35rem;
                    }
                }
            }
        }
    }

    &:hover .v-chart-controls {
        visibility: visible;
    }

    .v-message {
        color: var(--gray);
        font-size: 1rem;
    }

    ::v-deep(.chart-tooltip) {
        z-index: 700;
    }
}
</style>
