<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-message class="half" type="info" v-if="!variables?.length">
        {{ i18n.t("messages.variables.none") }}
    </v-message>

    <v-table class="half" v-model:value="variables" v-else>
        <template v-slot:header>
            <v-table-cell>{{ i18n.t("labels.name._") }}</v-table-cell>
            <v-table-cell grow>{{ i18n.t("labels.properties") }}</v-table-cell>
            <v-table-cell></v-table-cell>
        </template>

        <template v-slot="variable">
            <v-table-cell>
                {{ variable.value.name }}
            </v-table-cell>

            <v-table-cell grow>
                <template v-if="variableDefined(variable.value)">
                    <v-labels
                        :labels="{dynamic: true, label: variable.value.label}"
                        v-if="variable.value.dynamic"
                    ></v-labels>

                    <v-labels :labels="{dynamic: false, value: variable.value.value}" v-else></v-labels>
                </template>

                <span class="not-defined" v-else>
                    {{ i18n.t("messages.notDefined") }}
                </span>
            </v-table-cell>

            <v-table-cell>
                <v-button
                    class="reveal"
                    icon="pencil-alt"
                    @click="editVariable(variable.index)"
                    v-tooltip="i18n.t('labels.variables.edit')"
                ></v-button>

                <v-button
                    class="reveal"
                    icon="eraser"
                    :disabled="!variableDefined(variable.value)"
                    @click="clearVariable(variable.index)"
                    v-tooltip="i18n.t('labels.variables.clear')"
                ></v-button>
            </v-table-cell>
        </template>
    </v-table>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import natsort from "natsort";
import {useI18n} from "vue-i18n";
import {SetupContext, computed} from "vue";

import {TemplateVariable} from "types/api";

import {useUI} from "@/components/ui";

import {ModalTemplateVariableParams} from "../modal/template-variable.vue";

export default {
    inheritAttrs: false,
    props: {
        parsed: {
            required: true,
            type: Array as () => Array<TemplateVariable>,
        },
        value: {
            required: true,
            type: Array as () => Array<TemplateVariable>,
        },
    },
    setup(props: Record<string, any>, ctx: SetupContext): Record<string, unknown> {
        const i18n = useI18n();
        const ui = useUI();

        const available = computed(
            (): Array<TemplateVariable> => {
                const defined = props.value.map((variable: TemplateVariable) => variable.name);
                return props.parsed.filter((variable: TemplateVariable) => !defined.includes(variable.name));
            },
        );

        const variables = computed(
            (): Array<TemplateVariable> => {
                const sorter = natsort();
                return cloneDeep(props.value.concat(available.value)).sort((a: TemplateVariable, b: TemplateVariable) =>
                    sorter(a.name, b.name),
                );
            },
        );

        const clearVariable = (index: number): void => {
            ctx.emit("update:value", props.value.slice(0, index).concat(props.value.slice(index + 1)));
        };

        const editVariable = async (index: number): Promise<void> => {
            const variable = await ui.modal<TemplateVariable>("template-variable", {
                available: available.value,
                variable: cloneDeep(variables.value[index]),
            } as ModalTemplateVariableParams);

            if (variable) {
                index = props.value.findIndex((v: TemplateVariable) => v.name === variable.name);
                if (index !== -1) {
                    ctx.emit(
                        "update:value",
                        props.value.slice(0, index).concat(variable, props.value.slice(index + 1)),
                    );
                }
            }
        };

        const variableDefined = (variable: TemplateVariable): boolean => {
            return Boolean((!variable.dynamic && variable.value) || (variable.dynamic && variable.label));
        };

        return {
            clearVariable,
            editVariable,
            i18n,
            variableDefined,
            variables,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-table {
    .not-defined {
        color: var(--light-gray);
        text-transform: lowercase;
    }
}
</style>
