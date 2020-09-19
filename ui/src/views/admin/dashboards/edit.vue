<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content>
        <teleport to="body">
            <v-modal-template-variable></v-modal-template-variable>
        </teleport>

        <v-toolbar clip="content">
            <v-button
                icon="save"
                :disabled="erred || loading || saving"
                @click="saveDashboard(true)"
                v-if="prevRoute.name === 'dashboards-show' && !template && modifiers.alt"
            >
                {{ i18n.t("labels.saveAndGo") }}
            </v-button>

            <v-button icon="save" :disabled="erred || loading || saving" @click="saveDashboard(false)" v-else>
                {{ i18n.t(`labels.${template ? "templates" : "dashboards"}.save`) }}
            </v-button>

            <v-button icon="trash" @click="deleteDashboard()" v-if="!erred && edit && modifiers.alt">
                {{ i18n.t("labels.delete") }}
            </v-button>

            <v-button :disabled="erred" @click="redirectPrev()" v-else>
                {{ i18n.t("labels.cancel") }}
            </v-button>

            <v-divider vertical></v-divider>

            <v-button icon="undo" :disabled="erred || loading || !routeGuarded" @click="reset()">
                {{ i18n.t("labels.reset") }}
            </v-button>

            <v-divider vertical></v-divider>

            <template v-if="dashboard && edit">
                <v-spacer></v-spacer>

                <v-label class="note" v-if="dashboard.modifiedAt">
                    {{ i18n.t("messages.lastModified", [formatDate(dashboard.modifiedAt, i18n.t("date.long"))]) }}
                </v-label>
            </template>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <v-message-error v-else-if="erred"></v-message-error>

        <template v-else-if="dashboard">
            <h1 v-if="!$route.params.section">{{ i18n.t("labels.general") }}</h1>

            <v-form ref="form" class="third" v-show="!$route.params.section">
                <v-label>{{ i18n.t("labels.name._") }}</v-label>
                <v-input
                    required
                    :custom-validity="objectNameValidity('dashboards', dashboard.id)"
                    :delay="350"
                    :help="i18n.t('help.dashboards.name')"
                    :pattern="namePattern"
                    :placeholder="i18n.t('labels.name.choose')"
                    v-autofocus.select
                    v-model:value="dashboard.name"
                ></v-input>

                <template v-if="link">
                    <v-label>{{ i18n.t("labels.templates._", 1) }}</v-label>
                    <v-flex class="columns">
                        <v-select
                            required
                            :options="templates"
                            :placeholder="i18n.t('labels.templates.select')"
                            v-model:value="dashboard.link"
                        >
                            <template v-slot:dropdown-placeholder v-if="templates.length === 0">
                                <v-label>{{ i18n.t("messages.templates.none") }}</v-label>
                            </template>
                        </v-select>

                        <v-button
                            icon="pencil-alt"
                            :to="{name: 'admin-dashboards-edit', params: {id: String(dashboard.link)}}"
                            :style="{visibility: dashboard.link ? 'visible' : 'hidden'}"
                        >
                            {{ i18n.t("labels.templates.edit") }}
                        </v-button>
                    </v-flex>
                </template>

                <template v-else>
                    <v-label>{{ i18n.t("labels.title") }}</v-label>
                    <v-input
                        :delay="350"
                        :help="i18n.t('help.dashboards.title')"
                        v-model:value="dashboard.options.title"
                    ></v-input>
                </template>
            </v-form>

            <template v-if="$route.params.section === 'layout'">
                <v-grid
                    :readonly="link"
                    @add-item="!link ? addItem() : undefined"
                    v-model:layout="resolvedDashboard.layout"
                    v-model:value="resolvedDashboard.items"
                >
                    <template v-slot="item">
                        <v-chart
                            :legend="item.value.options.legend"
                            @click="!link ? editItem(item.index) : undefined"
                            v-model:value="dashboardRefs[`chart|${item.value.options.id}`]"
                            v-if="item.value.type === 'chart'"
                        >
                        </v-chart>

                        <v-text v-model:value="item.value.options" v-else-if="item.value.type === 'text'"></v-text>
                    </template>
                </v-grid>
            </template>

            <template v-else-if="$route.params.section === 'variables' && variables?.length">
                <h1>{{ i18n.t("labels.variables._") }}</h1>

                <v-form-template-variables
                    :parsed="variables"
                    v-model:value="dashboard.options.variables"
                ></v-form-template-variables>
            </template>
        </template>
    </v-content>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import {computed, onBeforeMount, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {onBeforeRouteLeave, onBeforeRouteUpdate, useRouter} from "vue-router";
import {useStore} from "vuex";

import {APIResponse, Chart, Dashboard, TemplateVariable} from "types/api";
import {FormComponent, SelectOption} from "types/ui";

import common, {namePattern} from "@/common";
import {ModalConfirmParams} from "@/components/modal/confirm.vue";
import {useUI} from "@/components/ui";
import {formatDate} from "@/helpers/date";
import {objectNameValidity} from "@/helpers/validity";
import api from "@/lib/api";
import {renderChart, renderDashboard, resolveDashboardReferences} from "@/lib/objects";
import {State} from "@/store";

import FormTemplateVariablesComponent from "../common/form/template-variables.vue";
import ModalTemplateVariableComponent from "../common/modal/template-variable.vue";

const defaultDashboard: Dashboard = {
    id: "",
    name: "",
    options: {
        title: "",
        variables: [],
    },
    layout: {
        columns: 1,
        rowHeight: 260,
        rows: 1,
    },
    items: [],
};

const defaultDashboardLinked: Dashboard = {
    id: "",
    name: "",
    options: {
        variables: [],
    },
};

export default {
    components: {
        "v-form-template-variables": FormTemplateVariablesComponent,
        "v-modal-template-variable": ModalTemplateVariableComponent,
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
            onBulkRejected,
            onFetchRejected,
            prevRoute,
            routeGuarded,
            watchGuard,
            unwatchGuard,
        } = common;

        const dashboard = ref<Dashboard | null>(null);
        const dashboardRefs = ref<Record<string, unknown>>({});
        const data = ref<Record<string, string>>({});
        //     const dynamicData = ref<Record<string, Array<string>>>({});
        const form = ref<FormComponent | null>(null);
        const invalid = ref(false);
        const linked = ref<Dashboard | null>(null);
        const saving = ref(false);
        const templates = ref<Array<SelectOption>>([]);
        const variables = ref<Array<TemplateVariable>>([]);

        //     const dynamicOptions = computed(
        //         (): Record<string, Array<SelectOption>> => {
        //             return Object.keys(dynamicData.value).reduce(
        //                 (out: Record<string, Array<SelectOption>>, name: string) => {
        //                     out[name] = dynamicData.value[name].map(value => ({label: value, value}));
        //                     return out;
        //                 },
        //                 {},
        //             );
        //         },
        //     );

        //     const dynamicVariables = computed(
        //         (): Array<TemplateVariable> => {
        //             return dashboard.value?.options?.variables?.filter(variable => variable.dynamic) ?? [];
        //         },
        //     );

        const edit = computed(
            () => router.currentRoute.value.params.id !== "new" && router.currentRoute.value.params.id !== "link",
        );

        const link = computed(() => router.currentRoute.value.params.id === "link" || Boolean(dashboard.value?.link));

        const resolvedDashboard = computed((): Dashboard | null => {
            if (linked.value) {
                return renderDashboard(linked.value, data.value);
            } else if (!link.value) {
                return dashboard.value;
            }

            return null;
        });

        const template = computed(() => !link.value && variables.value.length > 0);

        const addItem = (): void => {
            // console.debug("addItem!");
        };

        const deleteDashboard = async (): Promise<void> => {
            if (dashboard.value === null) {
                throw Error("cannot get dashboard");
            }

            const ok = await ui.modal<boolean>("confirm", {
                button: {
                    label: i18n.t(`labels.dashboards.delete`, 1),
                    danger: true,
                },
                message: i18n.t(`messages.dashboards.delete`, dashboard.value, 1),
            } as ModalConfirmParams);

            if (ok) {
                api.delete("dashboards", dashboard.value.id).then(() => {
                    ui.notify(i18n.t(`messages.dashboards.deleted`, 1), "success");
                    unwatchGuard();
                    router.push({name: "admin-dashboards-list"});
                }, onFetchRejected);
            }
        };

        const redirect = (go: boolean): void => {
            router.push(
                go || prevRoute.value?.name === "dashboards-show" || router.currentRoute.value.query.from === "basket"
                    ? {
                          name: "dashboards-show",
                          params: {id: dashboard.value?.name as string},
                          query: prevRoute.value?.query,
                      }
                    : {name: "admin-dashboards-list", query: template.value ? {kind: "template"} : {}},
            );
        };

        const redirectPrev = (): void => {
            router.push(
                prevRoute.value ?? {name: "admin-dashboards-list", query: template.value ? {kind: "template"} : {}},
            );
        };

        const reset = async (force = false): Promise<void> => {
            if (!force) {
                const ok = await ui.modal<boolean>("confirm", {
                    button: {
                        label: i18n.t("labels.dashboards.reset"),
                        danger: true,
                    },
                    message: i18n.t("messages.unsavedLost"),
                } as ModalConfirmParams);

                if (!ok) {
                    return;
                }
            }

            let promise: Promise<APIResponse<Dashboard | null>>;
            if (edit.value) {
                promise = api.object<Dashboard>("dashboards", router.currentRoute.value.params.id as string);
            } else {
                promise = Promise.resolve(
                    router.currentRoute.value.query.from === "basket"
                        ? {
                              data: {
                                  id: "",
                                  name: "",
                                  items: cloneDeep(store.state.basket),
                              },
                          }
                        : {
                              data:
                                  link.value && router.currentRoute.value.query.template
                                      ? ({link: router.currentRoute.value.query.template} as Dashboard)
                                      : null,
                          },
                );
            }

            form.value?.reset();
            invalid.value = false;
            store.commit("loading", true);

            promise
                .then(async response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get dashboard");
                    }

                    if (router.currentRoute.value.params.id === "link" || response.data?.link) {
                        await api
                            .objects<Dashboard>("dashboards", {kind: "template"})
                            .then(response => {
                                if (response.data) {
                                    templates.value = response.data.map(dashboard => ({
                                        label: dashboard.name,
                                        value: dashboard.id,
                                    }));
                                }
                            });

                        dashboard.value = merge({}, defaultDashboardLinked, response.data);
                    } else {
                        dashboard.value = merge({}, defaultDashboard, response.data);
                    }

                    await resolveDashboardReferences(dashboard.value?.items ?? []).then(response => {
                        response.forEach(ref => {
                            switch (ref.type) {
                                case "chart": {
                                    const chart = ref.value as Chart;
                                    dashboardRefs.value[`chart|${chart.id}`] = renderChart(chart, data.value);
                                    break;
                                }
                            }
                        });
                    }, onBulkRejected);

                    watchGuard(dashboard);
                }, onFetchRejected)
                .finally(() => {
                    updateRouteData();
                    store.commit("loading", false);
                });
        };

        const saveDashboard = (go: boolean): void => {
            if (dashboard.value === null) {
                throw Error("cannot get dashboard");
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

            const obj: Dashboard = cloneDeep(dashboard.value);
            obj.template = template.value;

            api.saveObject<Dashboard>("dashboards", obj).then(() => {
                ui.notify(i18n.t("messages.dashboards.saved"), "success");

                // Successfuly saved and data came from basket, thus empty it
                if (router.currentRoute.value.query.from === "basket") {
                    store.commit("basket", []);
                }

                unwatchGuard();
                redirect(go);
            }, onFetchRejected);
        };

        const updateRouteData = (clear = false): void => {
            store.commit(
                "routeData",
                !clear
                    ? {
                          invalid: {
                              general: invalid.value,
                          },
                          link: link.value,
                          items: dashboard.value?.items?.length ?? 0,
                          variables: variables.value?.length ?? 0,
                      }
                    : null,
            );
        };

        onBeforeRouteLeave(beforeRoute);

        onBeforeRouteUpdate(beforeRoute);

        onBeforeMount(() => applyRouteParams());

        onMounted(() => {
            ui.title(`${i18n.t("labels.dashboards._", 2)} – ${i18n.t("labels.adminPanel")}`);

            reset(true);
        });

        onBeforeUnmount(() => updateRouteData(true));

        //     watch(
        //         dashboard,
        //         async to => {
        //             if (!to) {
        //                 return;
        //             }

        //             if (to.link) {
        //                 // Skip loading linked dashboard if didn't change
        //                 if (to.link !== linked.value?.id) {
        //                     await api.object<Dashboard>("dashboards", to.link).then(response => {
        //                         if (response.data === undefined) {
        //                             return Promise.reject("cannot get linked dashboard");
        //                         }

        //                         linked.value = response.data;
        //                         variables.value = parseDashboardVariables(response.data);
        //                     });
        //                 }
        //             } else {
        //                 variables.value = parseDashboardVariables(to);
        //             }

        //             dynamicData.value = await resolveVariables(to.options?.variables ?? []);

        //             data.value = Object.keys(dynamicData.value).reduce((out: Record<string, string>, key: string) => {
        //                 out[key] = dynamicData.value[key]?.[0];
        //                 return out;
        //             }, {});

        //             updateRouteData();
        //         },
        //         {deep: true},
        //     );

        watch(
            () => router.currentRoute.value.params,
            (to, from) => {
                if (to.id !== from.id) {
                    reset(true);
                } else if (to.section !== from.section) {
                    invalid.value = !form.value?.checkValidity();
                    updateRouteData();
                }
            },
        );

        return {
            addItem,
            dashboard,
            dashboardRefs,
            //         colors,
            //         data,
            deleteDashboard,
            //         dynamicOptions,
            //         dynamicVariables,
            edit,
            erred,
            form,
            formatDate,
            i18n,
            link,
            loading,
            modifiers,
            namePattern,
            objectNameValidity,
            prevRoute,
            redirectPrev,
            reset,
            resolvedDashboard,
            routeGuarded,
            saveDashboard,
            saving,
            template,
            templates,
            variables,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";
@import "../../../components/ui/components/form/mixins";

.v-content {
    @include content;

    .columns {
        @include form;
    }
}
</style>
