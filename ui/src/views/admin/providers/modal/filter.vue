<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="provider-filter" :title="i18n.t(`labels.filters.${edit ? 'edit' : 'add'}`)" @show="onShow">
        <template v-slot="modal">
            <v-form>
                <v-label>{{ i18n.t("labels.filters.action._") }}</v-label>
                <v-select
                    :help="i18n.t('help.filters.action')"
                    :options="actions"
                    :placeholder="i18n.t('labels.filters.action.select')"
                    v-autofocus
                    v-model:value="rule.action"
                ></v-select>

                <v-label>{{ i18n.t("labels.labels._", 1) }}</v-label>
                <v-input
                    :help="i18n.t('help.filters.label')"
                    :placeholder="i18n.t('labels.placeholders.example', [NameLabel])"
                    v-model:value="rule.label"
                ></v-input>

                <v-label>{{ i18n.t("labels.filters.pattern") }}</v-label>
                <v-input :help="i18n.t('help.filters.pattern')" v-model:value="rule.pattern"></v-input>

                <template v-if="rule.action === 'relabel'">
                    <v-label>{{ i18n.t("labels.filters.targets._") }}</v-label>
                    <v-flex class="target" :key="index" v-for="(target, index) in targets">
                        <v-input
                            :delay="350"
                            :placeholder="i18n.t('labels.labels._', 1)"
                            :ref="index === targets.length - 1 ? 'label' : undefined"
                            v-model:value="target.key"
                        ></v-input>

                        <v-input
                            :delay="350"
                            :placeholder="i18n.t('labels.value')"
                            v-model:value="target.value"
                        ></v-input>

                        <v-button icon="times" @click="removeTarget(index)"></v-button>
                    </v-flex>

                    <v-button class="add" icon="plus" :disabled="'' in rule.targets" @click="addTarget">
                        {{ i18n.t("labels.filters.targets.add") }}
                    </v-button>
                </template>

                <template v-else-if="rule.action === 'rewrite'">
                    <v-label>{{ i18n.t("labels.filters.into") }}</v-label>
                    <v-input :help="i18n.t('help.filters.into')" v-model:value="rule.into"></v-input>
                </template>

                <template v-slot:bottom>
                    <v-button @click="modal.close(false)">
                        {{ i18n.t("labels.cancel") }}
                    </v-button>

                    <v-button :disabled="!rule.action || !rule.label" primary @click="modal.close(rule)">
                        {{ i18n.t("labels.filters.set") }}
                    </v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import {computed, nextTick, ref, watch} from "vue";
import {useI18n} from "vue-i18n";

import {FilterRule} from "types/api";
import {SelectOption} from "types/ui";

import {NameLabel} from "@/lib/labels";

export interface ModalProviderFilterParams {
    edit: boolean;
    rule: FilterRule;
}

interface Target {
    key: string;
    value: string;
}

const defaultRule: FilterRule = {
    action: "discard",
    label: "",
    pattern: "",
    into: "",
    targets: {},
};

const ruleActions = ["discard", "relabel", "rewrite", "sieve"];

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        const edit = ref(false);
        const label = ref<HTMLElement | null>(null);
        const rule = ref<FilterRule>(cloneDeep(defaultRule));
        const targets = ref<Array<Target>>([]);

        const actions = computed<Array<SelectOption>>(() =>
            ruleActions.map(action => ({label: action, value: action})),
        );

        const addTarget = (): void => {
            targets.value.push({key: "", value: ""});
            nextTick(() => label.value?.focus());
        };

        const onShow = (params: ModalProviderFilterParams): void => {
            edit.value = params.edit;
            rule.value = merge({}, defaultRule, params.rule);
        };

        const removeTarget = (index: number): void => {
            targets.value.splice(index, 1);
        };

        watch(
            targets,
            to => {
                rule.value.targets = to.reduce((out: Record<string, string>, target: Target) => {
                    out[target.key] = target.value;
                    return out;
                }, {});
            },
            {deep: true},
        );

        return {
            actions,
            addTarget,
            edit,
            i18n,
            label,
            NameLabel,
            onShow,
            removeTarget,
            rule,
            targets,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../../mixins";

.v-modal {
    ::v-deep(.v-modal-content) {
        @include content;

        .v-button.add {
            margin-top: 0.5rem;
            width: 100%;
        }

        .target {
            .v-input {
                min-width: auto;
                margin-right: 0.5rem;
            }

            & + .target {
                margin-top: 0.35rem;
            }
        }
    }
}
</style>
