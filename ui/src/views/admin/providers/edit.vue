<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content>
        <teleport to="body">
            <v-modal-provider-filter></v-modal-provider-filter>
        </teleport>

        <v-toolbar clip="content">
            <v-button
                icon="save"
                :disabled="erred || loading || saving || testing"
                @click="saveProvider(true)"
                v-if="modifiers.alt"
            >
                {{ i18n.t("labels.saveAndGo") }}
            </v-button>

            <v-button icon="save" :disabled="erred || loading || saving || testing" @click="saveProvider(false)" v-else>
                {{ i18n.t("labels.providers.save") }}
            </v-button>

            <v-button icon="trash" @click="deleteProvider()" :disabled="testing" v-if="!erred && edit && modifiers.alt">
                {{ i18n.t("labels.delete") }}
            </v-button>

            <v-button :disabled="erred" :to="{name: 'admin-providers-list'}" v-else>
                {{ i18n.t("labels.cancel") }}
            </v-button>

            <v-divider vertical></v-divider>

            <v-button icon="undo" :disabled="erred || loading || testing || !routeGuarded" @click="reset()">
                {{ i18n.t("labels.reset") }}
            </v-button>

            <v-divider vertical></v-divider>

            <v-button icon="clipboard-check" :disabled="erred || loading || saving || testing" @click="testProvider">
                {{ i18n.t("labels.providers.test") }}
            </v-button>

            <template v-if="provider && edit">
                <v-spacer></v-spacer>

                <v-label class="note" v-if="provider.modifiedAt">
                    {{ i18n.t("messages.lastModified", [formatDate(provider.modifiedAt, i18n.t("date.long"))]) }}
                </v-label>
            </template>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <v-message-error v-else-if="erred"></v-message-error>

        <template v-else-if="provider">
            <h1 v-if="!$route.params.section">{{ i18n.t("labels.general") }}</h1>

            <v-form ref="form" class="third" v-show="!$route.params.section">
                <v-label>{{ i18n.t("labels.name._") }}</v-label>
                <v-input
                    required
                    :custom-validity="objectNameValidity('providers', provider.id)"
                    :delay="350"
                    :help="i18n.t('help.providers.name')"
                    :pattern="namePattern"
                    :placeholder="i18n.t('labels.name.choose')"
                    v-autofocus.select
                    v-model:value="provider.name"
                ></v-input>

                <v-label>{{ i18n.t("labels.connectors._", 1) }}</v-label>
                <v-select
                    class="half"
                    required
                    :options="providers"
                    :placeholder="i18n.t('labels.connectors.select')"
                    v-model:value="provider.connector.type"
                ></v-select>

                <v-label>{{ i18n.t("labels.providers.pollInterval") }}</v-label>
                <v-input
                    class="half"
                    :help="i18n.t('help.providers.pollInterval')"
                    :placeholder="i18n.t('labels.placeholders.example', ['1m, 30m, 2h'])"
                    v-model:value="provider.pollInterval"
                ></v-input>

                <v-message type="error" v-if="supportFailed">
                    {{ i18n.t("messages.providers.supportFailed", [provider.connector.type]) }}
                </v-message>

                <component
                    :is="support"
                    :settings="provider.connector.settings"
                    v-else-if="provider.connector && support !== null"
                ></component>
            </v-form>

            <template v-if="$route.params.section === 'filters'">
                <h1>{{ i18n.t("labels.filters._", 2) }}</h1>

                <v-form class="half">
                    <v-message type="info" v-if="!provider.filters?.length">
                        {{ i18n.t("messages.filters.none") }}
                    </v-message>

                    <v-table draggable v-model:value="provider.filters" v-else>
                        <template v-slot:header>
                            <v-table-cell>
                                {{ i18n.t("labels.filters.action._") }}
                            </v-table-cell>

                            <v-table-cell grow>
                                {{ i18n.t("labels.properties") }}
                            </v-table-cell>

                            <v-table-cell></v-table-cell>
                        </template>

                        <template v-slot="item">
                            <v-table-cell>
                                {{ item.value.action }}
                            </v-table-cell>

                            <v-table-cell grow>
                                <v-flex>
                                    <v-labels :labels="{[item.value.label]: item.value.pattern}"></v-labels>

                                    <template v-if="item.value.action === 'relabel'">
                                        <v-icon icon="arrow-right"></v-icon>
                                        <v-labels :labels="item.value.targets"></v-labels>
                                    </template>

                                    <template v-else-if="item.value.action === 'rewrite'">
                                        <v-icon icon="arrow-right"></v-icon>
                                        <v-labels :labels="{[item.value.label]: item.value.into}"></v-labels>
                                    </template>
                                </v-flex>
                            </v-table-cell>

                            <v-table-cell>
                                <v-button
                                    class="reveal"
                                    icon="pencil-alt"
                                    @click="editFilter(item.index)"
                                    v-tooltip="i18n.t('labels.filters.edit')"
                                ></v-button>

                                <v-button
                                    class="reveal"
                                    icon="times"
                                    @click="removeFilter(item.index)"
                                    v-tooltip="i18n.t('labels.filters.remove')"
                                ></v-button>
                            </v-table-cell>
                        </template>
                    </v-table>

                    <v-toolbar>
                        <v-button icon="plus" @click="addFilter">{{ i18n.t("labels.filters.add") }}</v-button>

                        <v-spacer></v-spacer>

                        <v-message icon="question-circle" type="note">
                            {{ i18n.t("messages.labels.emptyDiscarded") }}
                        </v-message>
                    </v-toolbar>
                </v-form>
            </template>
        </template>
    </v-content>
</template>

<script lang="ts">
import merge from "lodash/merge";
import {
    Component,
    computed,
    defineAsyncComponent,
    onBeforeMount,
    onBeforeUnmount,
    onMounted,
    ref,
    shallowRef,
    watch,
} from "vue";
import {useI18n} from "vue-i18n";
import {onBeforeRouteLeave, onBeforeRouteUpdate, useRouter} from "vue-router";
import {useStore} from "vuex";

import {APIResponse, FilterRule, Provider} from "types/api";
import {FormComponent, SelectOption} from "types/ui";

import common, {namePattern} from "@/common";
import {ModalConfirmParams} from "@/components/modal/confirm.vue";
import {useUI} from "@/components/ui";
import {formatDate} from "@/helpers/date";
import {objectNameValidity} from "@/helpers/validity";
import api from "@/lib/api";
import {ProviderLabel} from "@/lib/labels";
import {State} from "@/store";

import ModalProviderFilterComponent, {ModalProviderFilterParams} from "./modal/filter.vue";

const defaultProvider: Provider = {
    id: "",
    name: "",
    connector: {
        type: "",
        settings: {},
    },
    filters: [],
    enabled: true,
};

const providers: Array<SelectOption> = [
    {label: "KairosDB", value: "kairosdb"},
    {label: "Prometheus", value: "prometheus"},
    {label: "RRDTool", value: "rrdtool"},
];

export default {
    components: {
        "v-modal-provider-filter": ModalProviderFilterComponent,
    },
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();
        const ui = useUI();

        const {
            applyRouteParams,
            beforeRoute,
            erred,
            loading,
            modifiers,
            onFetchRejected,
            routeGuarded,
            watchGuard,
            unwatchGuard,
        } = common;

        const form = ref<FormComponent | null>(null);
        const invalid = ref(false);
        const provider = ref<Provider | null>(null);
        const saving = ref(false);
        const support = shallowRef<Component | null>(null);
        const supportFailed = ref(false);
        const testing = ref(false);

        const edit = computed(() => router.currentRoute.value.params.id !== "new");

        const addFilter = async (): Promise<void> => {
            if (provider.value === null) {
                throw Error("cannot get provider");
            }

            const rule = await ui.modal<FilterRule>("provider-filter", {
                edit: false,
                rule: {},
            } as ModalProviderFilterParams);

            if (rule) {
                if (!provider.value.filters) {
                    provider.value.filters = [];
                }
                provider.value.filters.push(rule);
                updateRouteData();
            }
        };

        const deleteProvider = async (): Promise<void> => {
            if (provider.value === null) {
                throw Error("cannot get provider");
            }

            const ok = await ui.modal<boolean>("confirm", {
                button: {
                    label: i18n.t(`labels.providers.delete`, 1),
                    danger: true,
                },
                message: i18n.t(`messages.providers.delete`, provider.value, 1),
            } as ModalConfirmParams);

            if (ok) {
                api.delete("providers", provider.value.id).then(() => {
                    ui.notify(i18n.t(`messages.providers.deleted`, 1), "success");
                    unwatchGuard();
                    router.push({name: "admin-providers-list"});
                }, onFetchRejected);
            }
        };

        const editFilter = async (index: number): Promise<void> => {
            if (provider.value === null) {
                throw Error("cannot get provider");
            }

            const rule = await ui.modal<FilterRule>("provider-filter", {
                edit: false,
                rule: provider.value.filters?.[index],
            } as ModalProviderFilterParams);

            if (rule) {
                provider.value.filters?.splice(index, 1, rule);
                updateRouteData();
            }
        };

        const removeFilter = (index: number): void => {
            provider.value?.filters?.splice(index, 1);
            updateRouteData();
        };

        const redirect = (go: boolean): void => {
            router.push(
                go && provider.value?.name
                    ? {
                          name: "admin-metrics-list",
                          query: {filter: `{${ProviderLabel}=${JSON.stringify(provider.value.name)}}`},
                      }
                    : {name: "admin-providers-list"},
            );
        };

        const reset = async (force = false): Promise<void> => {
            if (!force) {
                const ok = await ui.modal<boolean>("confirm", {
                    button: {
                        label: i18n.t("labels.providers.reset"),
                        danger: true,
                    },
                    message: i18n.t("messages.unsavedLost"),
                } as ModalConfirmParams);

                if (!ok) {
                    return;
                }
            }

            let promise: Promise<APIResponse<Provider | null>>;
            if (edit.value) {
                promise = api.object<Provider>("providers", router.currentRoute.value.params.id as string);
            } else {
                promise = Promise.resolve({data: null});
            }

            form.value?.reset();
            invalid.value = false;
            store.commit("loading", true);

            promise
                .then(response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get provider");
                    }

                    provider.value = merge({}, defaultProvider, response.data);
                    watchGuard(provider);
                }, onFetchRejected)
                .finally(() => {
                    updateRouteData();
                    store.commit("loading", false);
                });
        };

        const saveProvider = (go: boolean): void => {
            if (provider.value === null) {
                throw Error("cannot get provider");
            }

            if (!routeGuarded) {
                redirect(go);
                return;
            } else if (!form.value?.checkValidity()) {
                ui.notify(i18n.t("messages.error.formVerify"), "error");
                invalid.value = true;
                updateRouteData();
                return;
            }

            api.saveObject<Provider>("providers", provider.value).then(() => {
                ui.notify(i18n.t("messages.providers.saved"), "success");
                unwatchGuard();
                redirect(go);
            }, onFetchRejected);
        };

        const testProvider = (): void => {
            if (provider.value === null) {
                throw Error("cannot get provider");
            }

            testing.value = true;

            api.testObject<Provider>("providers", provider.value)
                .then(response => {
                    if (response.error) {
                        ui.notify(i18n.t("messages.providers.test.error", [response.error]), "error");
                    } else {
                        ui.notify(i18n.t("messages.providers.test.success"), "success");
                    }
                })
                .finally(() => {
                    testing.value = false;
                });
        };

        const updateRouteData = (clear = false): void => {
            store.commit(
                "routeData",
                !clear
                    ? {
                          filters: provider.value?.filters?.length ?? 0,
                          invalid: {
                              general: invalid.value,
                          },
                      }
                    : null,
            );
        };

        onBeforeRouteLeave(beforeRoute);

        onBeforeRouteUpdate(beforeRoute);

        onBeforeMount(() => applyRouteParams());

        onMounted(() => {
            ui.title(`${i18n.t("labels.providers._", 2)} – ${i18n.t("labels.adminPanel")}`);

            reset(true);
        });

        onBeforeUnmount(() => updateRouteData(true));

        watch(
            () => provider.value?.connector.type,
            async to => {
                support.value = to
                    ? await defineAsyncComponent(() => import(/* webpackMode: "eager" */ `./form/${to}.vue`))
                    : null;
            },
            {deep: true},
        );

        watch(
            () => router.currentRoute.value.params.section,
            () => {
                if (!router.currentRoute.value.name?.toString().endsWith("-edit")) {
                    return;
                }

                invalid.value = !form.value?.checkValidity();
                updateRouteData();
            },
        );

        return {
            addFilter,
            deleteProvider,
            edit,
            editFilter,
            erred,
            form,
            formatDate,
            i18n,
            invalid,
            loading,
            modifiers,
            namePattern,
            objectNameValidity,
            provider,
            providers,
            removeFilter,
            reset,
            routeGuarded,
            saveProvider,
            saving,
            support,
            supportFailed,
            testing,
            testProvider,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";

.v-content {
    @include content;
}
</style>
