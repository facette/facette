<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="template-variable" :title="i18n.t('labels.variables.edit') + ` – ${variable.name}`" @show="onShow">
        <template v-slot="modal">
            <v-form>
                <v-tablist :tabs="tabs" v-model:value="variable.dynamic"></v-tablist>

                <template v-if="variable.dynamic">
                    <v-label>{{ i18n.t("labels.labels", 1) }}</v-label>
                    <v-input
                        :placeholder="i18n.t('labels.placeholders.example', ['instance'])"
                        v-autofocus.select
                        v-model:value="variable.label"
                    ></v-input>

                    <v-label>{{ i18n.t("labels.filters._", 1) }}</v-label>
                    <v-input
                        :placeholder="i18n.t('labels.placeholders.example', ['{__provider__=&quot;prometheus&quot;}'])"
                        v-model:value="variable.filter"
                    ></v-input>
                </template>

                <template v-else>
                    <v-label>{{ i18n.t("labels.value") }}</v-label>
                    <v-input
                        :placeholder="i18n.t('labels.placeholders.example', ['host.example.net'])"
                        v-autofocus.select
                        v-model:value="variable.value"
                    ></v-input>
                </template>

                <template v-slot:bottom>
                    <v-button @click="modal.close(false)">
                        {{ i18n.t("labels.cancel") }}
                    </v-button>

                    <v-button
                        :disabled="!variable.name || (variable.dynamic && !variable.label)"
                        primary
                        @click="modal.close(variable)"
                    >
                        {{ i18n.t("labels.variables.set") }}
                    </v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import {computed, ref} from "vue";
import {useI18n} from "vue-i18n";

import {TemplateVariable} from "types/api";
import {SelectOption, Tab} from "types/ui";

export interface ModalTemplateVariableParams {
    available: Array<TemplateVariable>;
    variable: TemplateVariable;
}

const defaultVariable: TemplateVariable = {
    name: "",
    value: "",
    label: "",
    filter: "",
    dynamic: false,
};

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        const available = ref<Array<TemplateVariable>>([]);

        const tabs = ref<Array<Tab>>([
            {label: i18n.t("labels.variables.fixed"), value: false},
            {label: i18n.t("labels.variables.dynamic"), value: true},
        ]);

        const variable = ref<TemplateVariable>(cloneDeep(defaultVariable));

        const variables = computed(
            (): Array<SelectOption> => available.value.map(variable => ({label: variable.name, value: variable.name})),
        );

        const onShow = (params: ModalTemplateVariableParams): void => {
            available.value = params.available;
            variable.value = merge({}, defaultVariable, params.variable);
        };

        return {
            i18n,
            onShow,
            tabs,
            variable,
            variables,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../../mixins";

.v-modal {
    ::v-deep(.v-modal-content) {
        @include content;
    }
}
</style>
