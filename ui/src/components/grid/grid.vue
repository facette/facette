<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div
        class="v-grid"
        ref="el"
        :aria-readonly="readonly"
        :style="`
            --columns: ${layout.columns || 1};
            --row-height: ${layout.rowHeight || 260}px;
            --rows: ${layout.rows || 1};
        `"
    >
        <template v-if="!readonly">
            <div class="v-grid-handle columns">
                <template v-if="layout.columns > 1">
                    <v-button
                        icon="minus"
                        :key="index"
                        :style="`grid-column-start: ${value}`"
                        @click="shrinkLayout('columns', index)"
                        @mouseenter="onHandleMouse($event, index)"
                        @mouseleave="onHandleMouse($event, index)"
                        v-for="(value, index) in layout.columns"
                        v-show="removable.x[index]"
                    ></v-button>
                </template>
            </div>

            <div class="v-grid-handle rows">
                <template v-if="layout.rows > 1">
                    <v-button
                        icon="minus"
                        :key="index"
                        :style="`grid-row-start: ${value}`"
                        @click="shrinkLayout('rows', index)"
                        @mouseenter="onHandleMouse($event, index)"
                        @mouseleave="onHandleMouse($event, index)"
                        v-for="(value, index) in layout.rows"
                        v-show="removable.y[index]"
                    ></v-button>
                </template>
            </div>

            <div class="v-grid-grow columns">
                <v-button icon="plus" @click="growLayout('rows')"></v-button>
            </div>

            <div class="v-grid-grow rows">
                <v-button icon="plus" @click="growLayout('columns')"></v-button>
            </div>
        </template>

        <div
            class="v-grid-container"
            ref="container"
            tabindex="-1"
            :class="{dragging: dragging !== null, hovering, resizing: resizing !== null, shrinking}"
            @click="!readonly && addItem()"
        >
            <v-grid-item placeholder :layout="placeholder" v-if="placeholder">
                <v-icon :icon="placeholderIcon" v-if="placeholderIcon"></v-icon>

                <v-label v-else>{{ placeholder.w }} x {{ placeholder.h }}</v-label>
            </v-grid-item>

            <v-grid-item
                :class="{highlight: index === highlightIndex}"
                :index="index"
                :key="index"
                :layout="item.layout"
                :readonly="readonly"
                @drag-item="!readonly && dragItem($event)"
                @remove-item="!readonly && removeItem(index)"
                @resize-item="!readonly && resizeItem(index)"
                v-for="(item, index) in value"
            >
                <slot v-bind="{index, value: item}"></slot>
            </v-grid-item>
        </div>
    </div>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import ResizeObserver from "resize-observer-polyfill";
import {SetupContext, computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from "vue";

import {GridItemLayout, GridLayout} from "types/api";

import common from "@/common";

interface GridItem {
    layout: GridItemLayout;
}

interface Position {
    x: number;
    y: number;
}

function compareGridItem(a: GridItem, b: GridItem) {
    const n = a.layout.y - b.layout.y;
    if (n !== 0) {
        return n;
    }

    return a.layout.x - b.layout.x;
}

export default {
    props: {
        highlightIndex: {
            default: null,
            type: Number,
        },
        layout: {
            required: true,
            type: Object as () => GridLayout,
        },
        readonly: {
            default: false,
            type: Boolean,
        },
        value: {
            required: true,
            type: Array as () => Array<GridItem>,
        },
    },
    setup(props: Record<string, any>, ctx: SetupContext): Record<string, unknown> {
        const {modifiers} = common;

        let columnWidth = 0;
        let domRect: DOMRect | null = null;
        let gridGap = 0;
        let position: Position | null = null;
        let resize: ResizeObserver | null = null;

        const container = ref<HTMLElement | null>(null);
        const dragging = ref<number | null>(null);
        const el = ref<HTMLElement | null>(null);
        const hovering = ref(false);
        const placeholder = ref<GridItemLayout | null>(null);
        const resizing = ref<number | null>(null);
        const shrinking = ref(false);

        const matrix = computed(
            (): Array<Array<number | null>> => {
                const out: Array<Array<number>> = Array.from({length: props.layout.rows}, () =>
                    Array(props.layout.columns).fill(null),
                );

                props.value.forEach((item: GridItem, index: number) => {
                    const layout: GridItemLayout = item.layout;

                    for (let y = layout.y; y < layout.y + layout.h; y++) {
                        for (let x = layout.x; x < layout.x + layout.w; x++) {
                            if (out[y]?.[x] === null) {
                                out[y][x] = index;
                            }
                        }
                    }
                });

                return out;
            },
        );

        const placeholderIcon = computed(() => {
            if (dragging.value !== null) {
                return "hand-pointer";
            } else if (hovering.value) {
                return "plus";
            } else if (shrinking.value) {
                return "times";
            }

            return null;
        });

        const removable = computed(() => {
            return {
                x: Array(props.layout.columns)
                    .fill(null)
                    .map((value, index) => matrix.value.filter(v => v[index] === null).length === props.layout.rows),
                y: Array(props.layout.rows)
                    .fill(null)
                    .map((value, index) => matrix.value[index].filter(v => v === null).length === props.layout.columns),
            };
        });

        const adaptLayoutRows = (value: Array<GridItem>): void => {
            const yDelta = Math.max(...value.map(item => item.layout.y + item.layout.h)) - props.layout.rows;
            if (yDelta > 0) {
                growLayout("rows", yDelta);
            }
        };

        const addItem = (): void => {
            if (position !== null) {
                ctx.emit("add-item", placeholder.value);
            }
        };

        const dragItem = (index: number | null): void => {
            if (el.value === null) {
                throw Error("cannot get element");
            }

            dragging.value = index;

            if (index !== null) {
                el.value.addEventListener("dragover", onDrag);
                el.value.addEventListener("drop", onDrag);
                placeholder.value = cloneDeep(props.value?.[index]?.layout);
            } else {
                el.value.removeEventListener("dragover", onDrag);
                el.value.removeEventListener("drop", onDrag);
                placeholder.value = null;
            }
        };

        const getCollisions = (value: Array<GridItem>, layout: GridItemLayout): Array<GridItem> => {
            return value
                .reduce((layouts: Array<GridItem>, item: GridItem) => {
                    if (
                        !(
                            (
                                item.layout === layout || // is same node
                                item.layout.x + item.layout.w <= layout.x || // is at left of node
                                item.layout.x >= layout.x + layout.w || // is at right of node
                                item.layout.y + item.layout.h <= layout.y || // is above node
                                item.layout.y >= layout.y + layout.h
                            ) // is below node
                        )
                    ) {
                        layouts.push(item);
                    }

                    return layouts;
                }, [])
                .sort(compareGridItem);
        };

        const getPosition = (ev: MouseEvent): Position | null => {
            if (domRect === null || !container.value?.contains(ev.target as Node)) {
                return null;
            }

            const x = Math.floor((ev.pageX - domRect.left + gridGap) / (columnWidth + gridGap));
            const y = Math.floor((ev.pageY - domRect.top + gridGap) / (props.layout.rowHeight + gridGap));

            return {
                x: x < 0 ? 0 : Math.min(x, props.layout.columns - 1),
                y: y < 0 ? 0 : Math.min(y, props.layout.rows - 1),
            };
        };

        const growLayout = (key: "columns" | "rows", delta = 1): void => {
            const layout: GridLayout = cloneDeep(props.layout);
            layout[key] += delta;

            ctx.emit("update:layout", layout);
        };

        const move = (value: Array<GridItem>, layout: GridItemLayout, x: number, y: number): void => {
            Object.assign(layout, {x, y});

            getCollisions(value, layout).forEach(item => move(value, item.layout, item.layout.x, layout.y + layout.h));
        };

        const onDrag = (ev: MouseEvent): void => {
            if (dragging.value === null) {
                return;
            }

            ev.preventDefault();

            switch (ev.type) {
                case "dragover": {
                    const pos = getPosition(ev);
                    if (pos === null) {
                        break;
                    }

                    const layout = props.value[dragging.value].layout;

                    const xDelta = props.layout.columns - (pos.x + layout.w);
                    if (xDelta < 0) {
                        pos.x += xDelta;
                    }

                    const yDelta = props.layout.rows - (pos.y + layout.h);
                    if (yDelta < 0) {
                        pos.y += yDelta;
                    }

                    if (pos.x !== placeholder.value?.x || pos.y !== placeholder.value.y) {
                        placeholder.value = {x: pos.x, y: pos.y, w: layout.w, h: layout.h};
                    }

                    break;
                }

                case "drop": {
                    const layout = props.value[dragging.value].layout;

                    // Only update layout if position has changed
                    if (
                        placeholder.value !== null &&
                        (placeholder.value.x !== layout.x || placeholder.value.y !== layout.y)
                    ) {
                        const value: Array<GridItem> = cloneDeep(props.value);

                        move(value, value[dragging.value].layout, placeholder.value.x, placeholder.value.y);
                        adaptLayoutRows(value);

                        ctx.emit("update:value", value.sort(compareGridItem));
                    }

                    break;
                }
            }
        };

        const onHandleMouse = (ev: MouseEvent, index: number): void => {
            if (ev.type === "mouseleave") {
                placeholder.value = null;
                shrinking.value = false;
                return;
            }

            if ((ev.target as HTMLElement).style.gridColumnStart !== "") {
                placeholder.value = {x: index, y: 0, w: 1, h: props.layout.rows};
            } else {
                placeholder.value = {x: 0, y: index, w: props.layout.columns, h: 1};
            }

            shrinking.value = true;
        };

        const onHighlightIndex = (to: number | null): void => {
            if (to !== null) {
                nextTick(() => document.getElementById(`item${to}`)?.scrollIntoView(true));
            }
        };

        const onLayout = (layout: GridLayout) => {
            columnWidth =
                el.value !== null ? (el.value.clientWidth - gridGap * (layout.columns + 1)) / layout.columns : 0;
        };

        const onMouse = (ev: MouseEvent): void => {
            switch (ev.type) {
                case "mouseleave": {
                    // Cursor has moved outside of the grid area, thus reset
                    // position.
                    position = null;
                    if (hovering.value) {
                        hovering.value = false;
                        placeholder.value = null;
                    }

                    break;
                }

                case "mousemove": {
                    const pos = getPosition(ev);
                    if (pos === null) {
                        position = null;
                        if (hovering.value) {
                            hovering.value = false;
                            placeholder.value = null;
                        }

                        break;
                    }

                    if (resizing.value !== null) {
                        const layout = props.value[resizing.value].layout;
                        const w = pos.x - layout.x + 1;
                        const h = pos.y - layout.y + 1;

                        if (
                            placeholder.value === null ||
                            (w !== placeholder.value.w && w >= 1) ||
                            (h !== placeholder.value.h && h >= 1)
                        ) {
                            placeholder.value = {x: layout.x, y: layout.y, w, h};
                        }

                        break;
                    }

                    // Only trigger hovering if current position doesn't match
                    // an existing grid cell
                    if (matrix.value?.[pos.y]?.[pos.x] === null) {
                        position = pos;
                        hovering.value = true;
                        updatePlaceholder();
                    } else {
                        hovering.value = false;
                        placeholder.value = null;
                    }

                    break;
                }

                case "mouseup": {
                    if (resizing.value === null || placeholder.value === null) {
                        break;
                    }

                    let layout = props.value[resizing.value].layout;

                    if (props.value[resizing.value].w !== placeholder.value.w || layout.h !== placeholder.value.h) {
                        const value: Array<GridItem> = cloneDeep(props.value);
                        layout = value[resizing.value].layout;
                        Object.assign(layout, {w: placeholder.value.w, h: placeholder.value.h});

                        getCollisions(value, layout).forEach(item =>
                            move(value, item.layout, item.layout.x, layout.y + layout.h),
                        );
                        adaptLayoutRows(value);

                        ctx.emit("update:value", value.sort(compareGridItem));
                    }

                    resizeItem(null);

                    break;
                }
            }
        };

        const removeItem = (index: number): void => {
            const value = cloneDeep(props.value);
            value.splice(index, 1);

            ctx.emit("update:value", value);
        };

        const resizeItem = (index: number | null): void => {
            if (el.value === null) {
                throw Error("cannot get element");
            }

            resizing.value = index;

            if (index !== null) {
                el.value.addEventListener("mouseup", onMouse);
                placeholder.value = cloneDeep(props.value?.[index]?.layout);
            } else {
                el.value.removeEventListener("mouseup", onMouse);
                placeholder.value = null;
            }
        };

        const shrinkLayout = (key: "columns" | "rows", index: number): void => {
            const value = cloneDeep(props.value);
            const layout = cloneDeep(props.layout);

            value.forEach((v: GridItem) => {
                if (key === "columns" && v.layout.x > index) {
                    v.layout.x--;
                } else if (key === "rows" && v.layout.y > index) {
                    v.layout.y--;
                }
            });

            if (key === "columns") {
                layout.columns--;
            } else if (key === "rows") {
                layout.rows--;
            }

            ctx.emit("update:value", value);
            ctx.emit("update:layout", layout);

            placeholder.value = null;
            shrinking.value = false;
        };

        const updatePlaceholder = () => {
            if (position === null) {
                return;
            }

            const {x, y} = position;

            if (modifiers.value.alt) {
                if (modifiers.value.shift) {
                    const boundaries = [y, y];
                    while (boundaries[0] > 0 && matrix.value?.[boundaries[0] - 1]?.[x] === null) {
                        boundaries[0]--;
                    }
                    while (boundaries[1] < props.layout.columns && matrix.value?.[boundaries[1] + 1]?.[x] === null) {
                        boundaries[1]++;
                    }
                    placeholder.value = {x, y: boundaries[0], w: 1, h: boundaries[1] - boundaries[0] + 1};
                } else {
                    const boundaries = [x, x];
                    while (boundaries[0] > 0 && matrix.value?.[y]?.[boundaries[0] - 1] === null) {
                        boundaries[0]--;
                    }
                    while (boundaries[1] < props.layout.columns && matrix.value?.[y]?.[boundaries[1] + 1] === null) {
                        boundaries[1]++;
                    }
                    placeholder.value = {x: boundaries[0], y, w: boundaries[1] - boundaries[0] + 1, h: 1};
                }
            } else {
                placeholder.value = {x, y, w: 1, h: 1};
            }
        };

        if (!props.readonly) {
            onMounted(() => {
                if (el.value === null) {
                    throw Error("cannot get element");
                } else if (container.value === null) {
                    throw Error("cannot get container");
                }

                el.value.addEventListener("mouseleave", onMouse);
                el.value.addEventListener("mousemove", onMouse);

                // Observe container resizing for DOMRect update, and get both
                // grid gap and initial column width (used by coordinate
                // computation).
                resize = new ResizeObserver(() => {
                    domRect = container.value?.getBoundingClientRect() ?? null;
                    onLayout(props.layout);
                });

                resize.observe(container.value);

                gridGap = parseFloat(getComputedStyle(container.value).rowGap);

                nextTick(() => onLayout(props.layout));
            });

            onBeforeUnmount(() => {
                if (el.value === null) {
                    throw Error("cannot get element");
                }

                el.value.removeEventListener("mouseleave", onMouse);
                el.value.removeEventListener("mousemove", onMouse);

                resize?.disconnect();
            });
        }

        watch(() => props.highlightIndex, onHighlightIndex, {immediate: true});

        watch(() => props.layout, onLayout, {deep: true});

        watch(modifiers, updatePlaceholder, {deep: true});

        return {
            addItem,
            container,
            dragging,
            dragItem,
            el,
            growLayout,
            hovering,
            onHandleMouse,
            placeholder,
            placeholderIcon,
            removable,
            removeItem,
            resizeItem,
            resizing,
            shrinking,
            shrinkLayout,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-grid {
    position: relative;

    &:not([aria-readonly="true"]) {
        margin: -1.125rem;
        padding: 1.125rem;
    }

    .v-grid-handle,
    .v-grid-grow {
        align-items: center;
        background-color: var(--toolbar-background);
        line-height: 1rem;
        position: absolute;

        &.columns {
            height: 1rem;
            left: 1.125rem;
            right: 1.125rem;
        }

        &.rows {
            bottom: 1.125rem;
            flex-direction: column;
            top: 1.125rem;
            width: 1rem;
        }

        .v-button {
            font-size: 0.65rem;
            height: inherit;
            margin: 0;
            min-width: auto;
            width: inherit;

            ::v-deep(.v-button-content) {
                border-radius: 0;
                color: var(--light-gray);
            }
        }
    }

    .v-grid-handle {
        display: grid;

        &.columns {
            column-gap: 0.65rem;
            grid-template-columns: repeat(var(--columns), 1fr);
            top: -0.5rem;
        }

        &.rows {
            grid-template-rows: repeat(var(--rows), var(--row-height));
            left: -0.5rem;
            row-gap: 0.65rem;

            .v-button {
                height: 100%;
            }
        }
    }

    .v-grid-grow {
        display: flex;

        &.columns {
            bottom: -0.5rem;
        }

        &.rows {
            right: -0.5rem;
        }

        .v-button {
            flex-grow: 1;
        }
    }

    .v-grid-container {
        column-gap: 0.65rem;
        display: grid;
        grid-template-columns: repeat(var(--columns), 1fr);
        grid-template-rows: repeat(var(--rows), var(--row-height));
        row-gap: 0.65rem;

        &:focus {
            outline: none;
        }

        &.dragging,
        &.hovering,
        &.resizing {
            .placeholder {
                background-color: rgba(var(--accent-rgb), 0.35);
                border-color: rgba(var(--accent-rgb), 0.65);
                color: var(--accent);
            }
        }

        &.shrinking .placeholder {
            background-color: rgba(211, 47, 47, 0.35);
            border: 0.15rem solid rgba(211, 47, 47, 0.65);
            color: var(--red);
        }
    }
}
</style>
