<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div
        class="v-modal"
        ref="el"
        role="dialog"
        tabindex="-1"
        @click="close(false)"
        @keydown.enter="onKeydown"
        @keydown.esc="close(false)"
        @keydown.tab="onKeydown"
        v-if="visible"
    >
        <div class="v-modal-content" :class="{title}" @click.stop>
            <h1 v-if="title">{{ title }}</h1>

            <slot v-bind="{close}"></slot>
        </div>
    </div>
</template>

<script lang="ts">
import {SetupContext, nextTick, onMounted, onUnmounted, ref} from "vue";

import {useUI} from "../..";

export default {
    props: {
        name: {
            required: true,
            type: String,
        },
        title: {
            default: null,
            type: String,
        },
    },
    setup(props: Record<string, any>, ctx: SetupContext): Record<string, unknown> {
        const ui = useUI();

        let resolve: ((value?: unknown) => void) | null = null;

        const el = ref<HTMLElement | null>(null);
        const visible = ref(false);

        const close = (value: unknown = null): void => {
            visible.value = false;
            ui.state.modals.current = null;
            resolve?.(value);
        };

        const focus = (): void => {
            el.value?.focus();
        };

        const open = (params: unknown = {}): Promise<unknown> => {
            // FIXME: hide previously existing overlays?
            visible.value = true;
            ctx.emit("show", params);

            nextTick(() => {
                if (el.value?.querySelector('[data-v-autofocus="true"]') === null) {
                    focus();
                }
            });

            return new Promise(fn => (resolve = fn));
        };

        const onKeydown = (ev: KeyboardEvent): void => {
            switch (ev.code) {
                case "Enter": {
                    if ((ev.target as HTMLElement).closest(".v-button, .v-form-bottom") === null) {
                        el.value
                            ?.querySelector('.v-button:is(.danger, .primary):not([aria-disabled="true"])')
                            ?.dispatchEvent(new Event("click"));
                    }

                    break;
                }

                case "Tab": {
                    const focusable = el.value?.querySelectorAll<HTMLElement>(
                        'input, textarea, [tabindex]:not([tabindex="-1"])',
                    );
                    if (!focusable) {
                        return;
                    }

                    let idx: number | null = null;
                    if (!ev.shiftKey && ev.target === focusable[focusable.length - 1]) {
                        idx = 0;
                    } else if (ev.shiftKey && ev.target === focusable[0]) {
                        idx = focusable.length - 1;
                    }

                    if (idx !== null) {
                        ev.preventDefault();
                        nextTick(() => focusable[idx as number].focus());
                    }

                    break;
                }
            }
        };

        onMounted(() => (ui.state.modals.entries[props.name] = {close, open}));

        onUnmounted(() => delete ui.state.modals.entries[props.name]);

        return {
            close,
            el,
            onKeydown,
            open,
            visible,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-modal {
    align-items: flex-start;
    background-color: var(--modal-background);
    bottom: 0;
    display: flex;
    justify-content: center;
    left: 0;
    position: fixed;
    right: 0;
    top: 0;
    z-index: 500;

    .v-modal-content {
        background-color: var(--background);
        border-radius: 0.2rem;
        box-shadow: 0 0.2rem 1rem var(--modal-content-shadow);
        margin: 10vh 0;
        max-height: 80vh;
        min-width: 25vw;
        overflow: auto;
        padding: 2rem;
        position: relative;

        h1 {
            background-color: var(--accent);
            border-radius: 0.2rem 0.2rem 0 0;
            color: white;
            font-size: 1rem;
            height: var(--toolbar-size);
            left: 0;
            line-height: var(--toolbar-size) !important;
            margin: 0;
            padding: 0 1.25rem;
            position: absolute;
            right: 0;
            top: 0;
        }

        &.title {
            padding-top: calc(2rem + var(--toolbar-size));
        }
    }

    ::v-deep(.v-form .v-form-bottom) {
        justify-content: flex-end;
        margin-top: 3rem;
    }
}
</style>
