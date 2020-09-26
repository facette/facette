<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div ref="el" class="v-tooltip" :class="{[state.anchor]: true, nowrap: state.nowrap}" v-if="state !== null">
        <v-markdown :content="state.content" v-if="state.content"></v-markdown>

        <span class="v-tooltip-shortcut" v-if="shortcutsEnabled && state.shortcut">
            <v-icon icon="keyboard"></v-icon>

            {{ shortcutLabel(state.shortcut) }}
        </span>
    </div>
</template>

<script lang="ts">
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from "vue";

import {TooltipState} from "types/ui";

import {useUI} from "../..";
import {shortcutLabel} from "../../directives/shortcut";
import {TooltipEvent} from "../../directives/tooltip";

export default {
    setup(): Record<string, unknown> {
        const ui = useUI();

        const el = ref<HTMLElement | null>(null);
        const state = ref<TooltipState | null>(null);

        const shortcutsEnabled = computed(() => ui.state.shortcuts.enabled);

        const onTooltip = ((ev: CustomEvent<TooltipEvent>): void => {
            state.value = ev.type === "tooltip-show" ? ev.detail.state : null;
        }) as (ev: Event) => void;

        onMounted(() => {
            document.addEventListener("tooltip-show", onTooltip);
            document.addEventListener("tooltip-hide", onTooltip);
        });

        onBeforeUnmount(() => {
            document.removeEventListener("tooltip-show", onTooltip);
            document.removeEventListener("tooltip-hide", onTooltip);
        });

        watch(state, to => {
            if (to === null) {
                el.value?.style.setProperty("visibility", null);
                return;
            }

            nextTick(() => {
                if (el.value === null) {
                    throw Error("cannot get element");
                }

                let left = 0;
                let top = 0;

                switch (to.anchor) {
                    case "bottom": {
                        left = to.domRect.left + to.domRect.width / 2 - el.value.clientWidth / 2;
                        top = to.domRect.top + to.domRect.height;
                        break;
                    }

                    case "left": {
                        left = to.domRect.left - el.value.clientWidth;
                        top = to.domRect.top + to.domRect.height / 2 - el.value.clientHeight / 2;
                        break;
                    }

                    case "right": {
                        left = to.domRect.left + to.domRect.width;
                        top = to.domRect.top + to.domRect.height / 2 - el.value.clientHeight / 2;
                        break;
                    }

                    case "top": {
                        left = to.domRect.left + to.domRect.width / 2 - el.value.clientWidth / 2;
                        top = to.domRect.top - el.value.clientHeight;
                        break;
                    }
                }

                if (left < 0) {
                    el.value.style.setProperty("--translate", `${left}px`);
                    left = 0;
                } else if (left + el.value.clientWidth > document.body.clientWidth) {
                    el.value.style.setProperty(
                        "--translate",
                        `${left + el.value.clientWidth - document.body.clientWidth}px`,
                    );
                    left = document.body.clientWidth - el.value.clientWidth;
                } else {
                    el.value.style.setProperty("--translate", "0");
                }

                Object.assign(el.value.style, {
                    left: `${left}px`,
                    top: `${top}px`,
                    visibility: "visible",
                });
            });
        });

        return {
            el,
            shortcutLabel,
            shortcutsEnabled,
            state,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-tooltip {
    background-color: var(--tooltip-background);
    border-radius: 0.2rem;
    box-shadow: 0 0.1rem 0.5rem var(--tooltip-shadow);
    cursor: default;
    display: flex;
    left: 0;
    max-width: 25vw;
    overflow-wrap: break-word;
    padding: 0.5rem 0.75rem;
    pointer-events: none;
    position: fixed;
    top: 0;
    visibility: hidden;
    will-change: left, top, transform;
    z-index: 700;

    &::before {
        border: 0.35rem solid transparent;
        content: "";
        display: block;
        height: 0;
        position: absolute;
        width: 0;
    }

    &.nowrap {
        white-space: nowrap;
    }

    &.bottom {
        transform: translateY(0.35rem);

        &::before {
            border-bottom-color: var(--tooltip-background);
            left: calc(50% - 0.35rem);
            top: -0.65rem;
            transform: translateX(var(--translate));
        }
    }

    &.left {
        transform: translateX(-0.35rem);

        &::before {
            border-left-color: var(--tooltip-background);
            right: -0.65rem;
            top: calc(50% - 0.35rem);
        }
    }

    &.right {
        transform: translateX(0.35rem);

        &::before {
            border-right-color: var(--tooltip-background);
            left: -0.65rem;
            top: calc(50% - 0.35rem);
        }
    }

    &.top {
        transform: translateY(-0.35rem);

        &::before {
            border-top-color: var(--tooltip-background);
            bottom: -0.65rem;
            left: calc(50% - 0.35rem);
            transform: translateX(var(--translate));
        }
    }

    :first-child {
        margin-top: 0;
    }

    :last-child {
        margin-bottom: 0;
    }

    .v-tooltip-shortcut {
        align-items: center;
        display: flex;
        font-size: 0.8rem;
        margin-left: 2rem;
        opacity: 0.425;

        &:first-child {
            margin-left: 0;
        }

        .v-icon {
            margin-right: 0.5rem;
        }
    }
}
</style>
