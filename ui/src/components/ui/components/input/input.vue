<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div
        class="v-input"
        ref="el"
        :aria-disabled="disabled || undefined"
        :aria-invalid="invalid || undefined"
        :aria-readonly="readonly || undefined"
        :class="{[type]: true, focus: focused}"
        @autofocus="focus($event.detail.select)"
        @click="focus()"
        @shortcut="!disabled && focus(true)"
    >
        <v-icon :icon="icon" @mousedown.prevent v-if="icon !== null"></v-icon>

        <v-label v-if="label">{{ label }}</v-label>

        <textarea
            ref="input"
            :disabled="disabled"
            :placeholder="placeholderLabel"
            :readonly="readonly"
            :required="required"
            :value="value"
            @focus="onFocus"
            @focusout="onFocus"
            @input="update(false)"
            @keydown.esc="onKeydownEsc"
            @report-validity="onReportValidity"
            @reset-validity="onResetValidity"
            v-if="type === 'textarea'"
        ></textarea>

        <input
            ref="input"
            :disabled="disabled"
            :pattern="pattern"
            :placeholder="placeholderLabel"
            :readonly="readonly"
            :required="required"
            :type="type === 'number' ? 'text' : 'type'"
            :value="value"
            @focus="onFocus"
            @focusout="onFocus"
            @input="update(false)"
            @keydown.esc="onKeydownEsc"
            @report-validity="onReportValidity"
            @reset-validity="onResetValidity"
            v-else
        />

        <v-icon class="help" icon="exclamation-triangle" @mousedown.prevent v-if="invalid"></v-icon>

        <v-icon class="help" icon="question-circle" @mousedown.prevent v-tooltip="help" v-else-if="help"></v-icon>
    </div>
</template>

<script lang="ts">
import {SetupContext, computed, ref} from "vue";

import {shortcutLabel} from "../../directives/shortcut";

export default {
    props: {
        customValidity: {
            default: null,
            type: Function,
        },
        delay: {
            default: 0,
            type: Number,
        },
        disabled: {
            default: false,
            type: Boolean,
        },
        help: {
            default: null,
            type: String,
        },
        icon: {
            default: null,
            type: String,
        },
        label: {
            default: null,
            type: String,
        },
        pattern: {
            default: null,
            type: String,
        },
        placeholder: {
            default: null,
            type: String,
        },
        readonly: {
            default: false,
            type: Boolean,
        },
        required: {
            default: false,
            type: Boolean,
        },
        type: {
            default: "text",
            type: String,
        },
        value: {
            default: "",
            required: true,
            type: [Number, String],
        },
    },
    setup(props: Record<string, any>, ctx: SetupContext): Record<string, unknown> {
        let pristine = true;
        let updateTimeout: number | null = null;

        const updateDebounce =
            props.delay > 0
                ? () => {
                      if (updateTimeout !== null) {
                          clearTimeout(updateTimeout);
                      }

                      updateTimeout = setTimeout(() => update(true), props.delay);
                  }
                : null;

        const el = ref<HTMLElement | null>(null);
        const focused = ref(false);
        const input = ref<HTMLInputElement | HTMLTextAreaElement | null>(null);
        const invalid = ref(false);

        const placeholderLabel = computed(() => {
            let placeholder = "";

            if (props.placeholder) {
                placeholder += props.placeholder;
            }

            if (el.value?.dataset.vShortcut) {
                if (placeholder) {
                    placeholder += " ";
                }
                placeholder += `(${shortcutLabel(el.value.dataset.vShortcut)})`;
            }

            return placeholder ?? null;
        });

        const clear = (): void => {
            input.value?.setCustomValidity("");

            ctx.emit("update:value", "");
            ctx.emit("clear");
        };

        const focus = (select = false): void => {
            if (!focused.value) {
                input.value?.[select ? "select" : "focus"]();
            }
        };

        const onFocus = async (ev: FocusEvent): Promise<void> => {
            focused.value = ev.type === "focus";
            if (!focused.value && pristine) {
                pristine = false;
            }
            if (!pristine) {
                onReportValidity();
            }
        };

        const onKeydownEsc = (): void => {
            if (props.type === "search") {
                if (props.value) {
                    clear();
                } else {
                    input.value?.blur();
                }
            }
        };

        const onReportValidity = (): void => {
            invalid.value = !input.value?.validity.valid;
        };

        const onResetValidity = (): void => {
            pristine = true;
            invalid.value = false;
        };

        const update = async (apply = false): Promise<void> => {
            if (apply || props.delay === 0) {
                const value = input.value?.value ?? "";

                if (props.customValidity !== null) {
                    input.value?.setCustomValidity(await props.customValidity(value));
                } else if (props.type === "number") {
                    input.value?.setCustomValidity(value && isNaN(Number(value)) ? "invalid number" : ""); // FIXME: make translatable?
                }

                if (pristine) {
                    pristine = false;
                }

                onReportValidity();

                ctx.emit("update:value", value ?? "");
            } else if (updateDebounce !== null) {
                updateDebounce();
            }
        };

        return {
            el,
            focus,
            focused,
            input,
            invalid,
            onFocus,
            onKeydownEsc,
            onReportValidity,
            onResetValidity,
            placeholderLabel,
            update,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "./mixins";

.v-input {
    @include input;

    input,
    textarea {
        background: none;
        border: none;
        color: inherit;
        flex-grow: 1;
        font: inherit;
        padding: 0;
        width: 0;

        &:focus,
        &:invalid {
            box-shadow: none;
            outline: none;
        }
    }

    input[type="number"] {
        -moz-appearance: textfield;

        &::-webkit-inner-spin-button,
        &::-webkit-outer-spin-button {
            -webkit-appearance: none;
        }
    }

    textarea {
        min-height: 6.25rem;
        tab-size: 4;
    }
}
</style>
