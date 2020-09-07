<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div class="v-editor monospace" :class="{focus: focused}" @autofocus="focus($event.detail.select)">
        <div
            class="v-editor-input"
            contenteditable="true"
            ref="input"
            spellcheck="false"
            @focus="onFocus"
            @focusout="onFocus"
            @keydown="onKeydown"
            @input="$emit('update:value', $event.target.innerText)"
        ></div>

        <v-highlight ref="highlight" :content="value"></v-highlight>
    </div>
</template>

<script lang="ts">
import {ComponentPublicInstance, nextTick, onBeforeUnmount, onMounted, ref, watch} from "vue";

export default {
    props: {
        value: {
            required: true,
            type: String,
        },
    },
    setup(props: Record<string, any>): Record<string, unknown> {
        const focused = ref(false);
        const highlight = ref<ComponentPublicInstance | null>(null);
        const input = ref<HTMLElement | null>(null);

        const focus = (select: boolean): void => {
            nextTick(() => {
                input.value?.focus();
                if (select) {
                    document.execCommand("selectAll");
                }
            });
        };

        const onFocus = async (ev: FocusEvent): Promise<void> => {
            focused.value = ev.type === "focus";
        };

        const onKeydown = (ev: KeyboardEvent): void => {
            if (ev.code !== "Enter" && ev.code !== "Tab") {
                return;
            }

            if (input.value === null) {
                throw Error("cannot get textarea");
            }

            ev.preventDefault();
            ev.stopPropagation();

            if (ev.code === "Enter") {
                document.execCommand("insertHTML", false, "<br>");
            } else {
                document.execCommand("insertText", false, "\t");
            }
        };

        const onScroll = (ev: Event): void => {
            if (highlight.value === null) {
                return;
            }

            highlight.value.$el.scrollTop = (ev.target as HTMLElement).scrollTop;
        };

        onMounted(() => {
            if (input.value === null) {
                throw Error("cannot find input");
            }

            input.value.addEventListener("scroll", onScroll);
            input.value.innerText = props.value;
        });

        onBeforeUnmount(() => {
            input.value?.removeEventListener("scroll", onScroll);
        });

        watch(
            () => props.value,
            to => {
                if (to === input.value?.innerText) {
                    return;
                }

                const sel = window.getSelection();
                if (sel !== null) {
                    let range = document.createRange();
                    range.selectNodeContents(input.value as Node);

                    sel.removeAllRanges();
                    sel.addRange(range);

                    document.execCommand("insertText", false, to);
                }
            },
        );

        return {
            focus,
            focused,
            onFocus,
            onKeydown,
            highlight,
            input,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-editor {
    background-color: var(--input-background);
    border-left: 0.25rem solid transparent;
    border-radius: 0.2rem;
    line-height: 1rem;
    padding: 0.65rem;
    position: relative;
    tab-size: 4;

    &.focus {
        border-left-color: var(--accent);
    }

    ::selection {
        color: transparent;
    }

    .v-editor-input {
        caret-color: var(--color);
        color: transparent;
        overflow: auto;
        resize: both;
        white-space: pre-wrap;
        width: 100%;
        word-break: break-all;

        &:focus {
            outline: none;
        }
    }

    .v-highlight {
        bottom: 0.65rem;
        left: 0.65rem;
        overflow: hidden;
        pointer-events: none;
        position: absolute;
        right: 0.65rem;
        top: 0.65rem;
        white-space: pre-wrap;
        word-break: break-all;

        ::v-deep(.v-highlight-line) {
            min-height: 1rem;
        }
    }
}
</style>
