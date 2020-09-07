<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <div class="v-highlight monospace" ref="el"></div>
</template>

<script lang="ts">
import {onMounted, ref, watch} from "vue";

import {Parser, TokenType} from "@/lib/parser";

function tokenClassName(type: string, prev?: TokenType, next?: TokenType): string | null {
    switch (type) {
        case TokenType.IDENT:
            switch (next) {
                case TokenType.LPAREN:
                    return "function";

                case TokenType.EQ:
                case TokenType.EQREGEXP:
                case TokenType.NEQ:
                case TokenType.NEQREGEXP:
                    return "label";

                default:
                    return [undefined, TokenType.COMMA, TokenType.LPAREN, TokenType.SPACE].includes(prev)
                        ? "name"
                        : null;
            }

        case TokenType.NUMBER:
        case TokenType.STRING:
            return type;
    }

    return null;
}

export default {
    props: {
        content: {
            required: true,
            type: String,
        },
    },
    setup(props: Record<string, any>): Record<string, unknown> {
        const el = ref<HTMLElement | null>(null);

        const update = (value: string): void => {
            if (el.value === null) {
                return;
            }

            const fragment = document.createDocumentFragment();
            const tokens = new Parser(value, true).tokens();

            let line = fragment.appendChild(document.createElement("div"));
            line.className = "v-highlight-line";

            tokens.forEach((tok, index) => {
                switch (tok.type) {
                    case TokenType.NEWLINE: {
                        line = fragment.appendChild(document.createElement("div"));
                        line.className = "v-highlight-line";
                        break;
                    }

                    default: {
                        const span = document.createElement("span");
                        const className = tokenClassName(
                            tok.type,
                            index > 1 ? tokens[index - 1].type : undefined,
                            tokens[index + 1]?.type,
                        );

                        if (className !== null) {
                            span.className = className;
                        }

                        span.appendChild(document.createTextNode(tok.text));

                        line.appendChild(span);
                    }
                }
            });

            el.value.textContent = "";
            el.value.append(...fragment.children);
        };

        onMounted(() => update(props.content));

        watch(
            () => props.content,
            to => update(to),
        );

        return {
            el,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-highlight {
    ::v-deep(.v-highlight-line) {
        color: var(--highlight-color);

        .function {
            color: var(--highlight-function-color);
        }

        .label {
            color: var(--highlight-label-color);
        }

        .name {
            color: var(--highlight-name-color);
        }

        .number {
            color: var(--highlight-number-color);
        }

        .string {
            color: var(--highlight-string-color);
        }
    }
}
</style>
