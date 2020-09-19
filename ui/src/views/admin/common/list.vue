<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-content>
        <v-toolbar clip="content">
            <v-button
                icon="plus"
                :to="{name: 'admin-providers-edit', params: {id: 'new'}}"
                v-shortcut="{keys: 'n', help: i18n.t(`labels.${type}.new`)}"
                v-if="type === 'providers'"
            >
                {{ i18n.t("labels.providers.new") }}
            </v-button>

            <v-button icon="plus" v-else>
                {{ i18n.t(`labels.${type}.new`) }}

                <template v-slot:dropdown>
                    <v-button
                        :to="{name: `admin-${type}-edit`, params: {id: 'new'}}"
                        v-shortcut="{keys: 'n', help: i18n.t(`labels.${type}.new`)}"
                    >
                        {{ i18n.t(`labels.${type}.new`) }}
                    </v-button>

                    <v-button
                        :to="{name: `admin-${type}-edit`, params: {id: 'link'}}"
                        v-shortcut="{keys: 'shift+n', help: i18n.t(`labels.templates.newFrom`)}"
                    >
                        {{ i18n.t("labels.templates.newFrom") }}
                    </v-button>
                </template>
            </v-button>

            <v-divider vertical></v-divider>

            <v-button icon="sync-alt" @click="getObjects" v-shortcut="{keys: 'r', help: i18n.t('labels.refresh.list')}">
                {{ i18n.t("labels.refresh._") }}
            </v-button>

            <template v-if="type == 'providers'">
                <v-divider vertical></v-divider>

                <v-button
                    icon="play"
                    :disabled="
                        selection.length === 0 || (selectionEnabled.length === 1 && selectionEnabled[0] === true)
                    "
                    @click="toggleProviders(selection, true)"
                >
                    {{ i18n.t("labels.providers.enable") }}
                </v-button>

                <v-button
                    icon="stop"
                    :disabled="
                        selection.length === 0 || (selectionEnabled.length === 1 && selectionEnabled[0] === false)
                    "
                    @click="toggleProviders(selection, false)"
                >
                    {{ i18n.t("labels.providers.disable") }}
                </v-button>

                <v-button
                    icon="arrow-alt-circle-down"
                    :disabled="
                        selection.length === 0 || (selectionEnabled.length === 1 && selectionEnabled[0] === false)
                    "
                    @click="pollProviders(selection)"
                >
                    {{ i18n.t("labels.providers.poll") }}
                </v-button>
            </template>

            <v-divider vertical></v-divider>

            <v-button
                icon="trash"
                :disabled="selection.length === 0"
                @click="deleteObjects(selection)"
                v-shortcut="{keys: 'd', help: i18n.t('labels.delete')}"
            >
                {{ i18n.t("labels.delete") }}
            </v-button>

            <v-spacer></v-spacer>

            <v-input
                icon="filter"
                type="search"
                :placeholder="i18n.t(`labels.${type}.filter`)"
                @clear="applyFilter"
                @focusout="applyFilter"
                @keypress.enter="applyFilter"
                v-model:value="filter"
                v-shortcut="{keys: 'f', help: i18n.t(`labels.${type}.filter`)}"
            ></v-input>
        </v-toolbar>

        <v-spinner v-if="loading"></v-spinner>

        <v-message-error @retry="getObjects" v-else-if="erred"></v-message-error>

        <template v-else>
            <h1>
                {{ i18n.t(`labels.${type}._`, 2) }}
                <span class="count" v-if="total">{{ total }}</span>
            </h1>

            <v-message class="selection" icon="clipboard-list" type="info" v-if="selection.length > 0">
                {{ i18n.t(`messages.${type}.selected`, [selection.length], selection.length) }}

                <v-button
                    icon="times-circle"
                    @click="clearSelection"
                    v-tooltip="i18n.t('labels.clearSelection')"
                ></v-button>
            </v-message>

            <v-tablist
                :tabs="tabs"
                @update:value="options.page = 1"
                v-model:value="options.kind"
                v-if="templatable"
            ></v-tablist>

            <v-message type="info" v-if="objects.length === 0">
                {{ i18n.t(`messages.${options.kind === "template" ? "templates" : type}.none`) }}
            </v-message>

            <template v-else>
                <v-table ref="table" selectable v-model:selection="selection" v-model:value="objects">
                    <template v-slot:header>
                        <v-table-cell grow>
                            <v-flex class="sort" @click="toggleSort('name')">
                                {{ i18n.t("labels.name._") }}
                                <v-icon
                                    :icon="`angle-${options.sort == '-name' ? 'down' : 'up'}`"
                                    v-if="options.sort == 'name' || options.sort == '-name'"
                                ></v-icon>
                            </v-flex>
                        </v-table-cell>

                        <v-table-cell>
                            <v-flex class="sort" @click="toggleSort('modifiedAt')">
                                {{ i18n.t("labels.lastModified") }}
                                <v-icon
                                    :icon="`angle-${options.sort == '-modifiedAt' ? 'down' : 'up'}`"
                                    v-if="options.sort == 'modifiedAt' || options.sort == '-modifiedAt'"
                                ></v-icon>
                            </v-flex>
                        </v-table-cell>

                        <v-table-cell></v-table-cell>
                    </template>

                    <template v-slot="obj">
                        <v-table-cell grow>
                            <v-flex>
                                <router-link
                                    class="link"
                                    :to="{name: `admin-${type}-edit`, params: {id: obj.value.id}}"
                                >
                                    {{ obj.value.name }}
                                </router-link>

                                <v-icon
                                    class="linked"
                                    icon="link"
                                    v-tooltip="i18n.t('labels.templates.instance')"
                                    v-if="templatable && obj.value.link"
                                ></v-icon>

                                <template v-else-if="type == 'providers'">
                                    <v-icon
                                        class="disabled"
                                        icon="stop-circle"
                                        v-tooltip="i18n.t('labels.providers.disabled')"
                                        v-if="!obj.value.enabled"
                                    ></v-icon>

                                    <v-icon
                                        class="error"
                                        icon="exclamation-circle"
                                        v-tooltip="i18n.t('messages.error._', [obj.value.error])"
                                        v-else-if="obj.value.error"
                                    ></v-icon>

                                    <v-icon
                                        class="enabled"
                                        icon="play-circle"
                                        v-tooltip="i18n.t('labels.providers.enabled')"
                                        v-else
                                    ></v-icon>
                                </template>
                            </v-flex>
                        </v-table-cell>

                        <v-table-cell>
                            {{ formatDate(obj.value.modifiedAt, i18n.t("date.long")) }}
                        </v-table-cell>

                        <v-table-cell>
                            <v-button class="icon" dropdown-anchor="right" icon="ellipsis-v">
                                <template v-slot:dropdown>
                                    <template v-if="type === 'providers'">
                                        <v-button
                                            icon="arrow-alt-circle-right"
                                            :to="{
                                                name: 'admin-metrics-list',
                                                query: {filter: `{__provider__=${JSON.stringify(obj.value.name)}}`},
                                            }"
                                        >
                                            {{ i18n.t("labels.goto.metrics") }}
                                        </v-button>

                                        <v-divider></v-divider>
                                    </template>

                                    <template v-else-if="templatable">
                                        <v-button
                                            icon="arrow-alt-circle-right"
                                            :to="{name: `${type}-show`, params: {id: obj.value.name}}"
                                            v-if="options.kind === 'plain'"
                                        >
                                            {{ i18n.t(`labels.goto.${type}`, 1) }}
                                        </v-button>

                                        <v-button
                                            icon="plus"
                                            :to="{
                                                name: `admin-${type}-edit`,
                                                params: {id: 'link'},
                                                query: {template: obj.value.id},
                                            }"
                                            v-else
                                        >
                                            {{ i18n.t("labels.templates.newFrom") }}
                                        </v-button>

                                        <v-divider></v-divider>
                                    </template>

                                    <v-button icon="clone" @click="cloneObject(obj.value)">
                                        {{ i18n.t("labels.clone") }}
                                    </v-button>

                                    <v-button icon="trash" @click="deleteObjects([obj.value])">
                                        {{ i18n.t("labels.delete") }}
                                    </v-button>

                                    <template v-if="type == 'providers'">
                                        <v-divider></v-divider>

                                        <v-button
                                            icon="play"
                                            @click="toggleProviders([obj.value], true)"
                                            v-if="!obj.value.enabled"
                                        >
                                            {{ i18n.t("labels.providers.enable") }}
                                        </v-button>

                                        <v-button icon="stop" @click="toggleProviders([obj.value], false)" v-else>
                                            {{ i18n.t("labels.providers.disable") }}
                                        </v-button>

                                        <v-button
                                            icon="sync-alt"
                                            :disabled="!obj.value.enabled"
                                            @click="pollProviders([obj.value])"
                                        >
                                            {{ i18n.t("labels.providers.poll") }}
                                        </v-button>
                                    </template>
                                </template>
                            </v-button>
                        </v-table-cell>
                    </template>
                </v-table>

                <v-paging :page-size="limit" :total="total" v-model:page="options.page"></v-paging>
            </template>
        </template>
    </v-content>
</template>

<script lang="ts">
import isEqual from "lodash/isEqual";
import {computed, onBeforeMount, onMounted, ref, unref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useRouter} from "vue-router";
import {useStore} from "vuex";

import {ListParams, ObjectBase, Provider} from "types/api";
import {Tab} from "types/ui";

import common from "@/common";
import {ModalConfirmParams} from "@/components/modal/confirm.vue";
import {ModalPromptParams} from "@/components/modal/prompt.vue";
import {useUI} from "@/components/ui";
import {formatDate} from "@/helpers/date";
import {objectNameValidity} from "@/helpers/validity";
import api from "@/lib/api";
import {State} from "@/store";

interface Options {
    filter: string;
    kind: string;
    page: number;
    sort: string;
}

const defaultOptions: Options = {
    filter: "",
    kind: "plain",
    page: 1,
    sort: "name",
};

const limit = 20;

export default {
    props: {
        type: {
            required: true,
            type: String,
        },
    },
    setup(props: Record<string, any>): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();
        const ui = useUI();

        const {erred, loading, onBulkRejected, onFetchRejected} = common;

        const filter = ref("");
        const objects = ref<Array<ObjectBase>>([]);
        const options = ref<Options>(Object.assign({}, defaultOptions));
        const selection = ref<Array<ObjectBase>>([]);
        const table = ref<HTMLTableElement | null>(null);

        const total = ref(0);

        const selectionEnabled = computed(() => {
            return props.type === "providers" ? (selection.value as Array<Provider>).map(obj => obj.enabled) : [];
        });

        const tabs = computed(
            (): Array<Tab> => [
                {label: i18n.t(`labels.${props.type}._`, 2), value: "plain"},
                {label: i18n.t("labels.templates._", 2), value: "template"},
            ],
        );

        const templatable = computed(() => props.type === "charts" || props.type === "dashboards");

        const applyFilter = (): void => {
            options.value.filter = filter.value;
        };

        const clearSelection = (): void => {
            selection.value = [];
        };

        const cloneObject = async (obj: ObjectBase): Promise<void> => {
            const name = await ui.modal<string | false>("prompt", {
                button: {
                    label: i18n.t("labels.clone"),
                    primary: true,
                },
                input: {
                    customValidity: objectNameValidity(props.type),
                    required: true,
                    value: `${obj.name}-clone`,
                },
                message: i18n.t(`labels.${props.type}.name`),
            } as ModalPromptParams);

            if (name !== false) {
                api.cloneObject(props.type, obj.id, {name}).then(() => getObjects());
            }
        };

        const deleteObjects = async (objects: Array<ObjectBase>): Promise<void> => {
            const ok = await ui.modal<boolean>("confirm", {
                button: {
                    label: i18n.t(`labels.${props.type}.delete`, objects.length),
                    danger: true,
                },
                message: i18n.t(
                    `messages.${props.type}.delete`,
                    objects.length > 1 ? {count: objects.length} : objects[0],
                    objects.length,
                ),
            } as ModalConfirmParams);

            if (ok) {
                api.bulk(
                    objects.map(obj => ({
                        endpoint: `/${props.type}/${obj.id}`,
                        method: "DELETE",
                    })),
                ).then(() => {
                    ui.notify(i18n.t(`messages.${props.type}.deleted`, objects.length), "success");
                    selection.value = [];
                    getObjects();
                }, onBulkRejected);
            }
        };

        const getObjects = (): void => {
            const listParams: ListParams = {
                limit,
                offset: (options.value.page - 1) * limit,
            };

            if (templatable.value && options.value.kind) {
                listParams.kind = options.value.kind;
            }

            if (options.value.sort) {
                listParams.sort = options.value.sort;
            }

            if (options.value.filter) {
                let parts: Array<string> = options.value.filter.split(" ");

                if (props.type === "providers") {
                    parts = parts.filter(part => !part.startsWith("enabled:"));
                    const enabled: string | undefined = parts.filter(part => part.startsWith("enabled:"))[0];
                    if (enabled) {
                        listParams.enabled = `${enabled.substr(8)}`;
                    }
                }

                if (parts.length > 0) {
                    listParams.name = `~(?:${parts.join("|")})`;
                }
            }

            store.commit("loading", true);

            api.objects(props.type, listParams)
                .then(response => {
                    if (response.data === undefined) {
                        return Promise.reject("cannot get objects");
                    }

                    const pagesCount: number | undefined = response.total
                        ? Math.ceil(response.total / limit)
                        : undefined;

                    // Switch back to first/last page if current empty
                    if (!response.data?.length && options.value.page > 1) {
                        options.value.page = pagesCount !== undefined ? pagesCount : 1;
                        return;
                    }

                    objects.value = response.data;
                    total.value = response.total ?? 0;
                }, onFetchRejected)
                .finally(() => {
                    store.commit("loading", false);
                });
        };

        const pollProviders = (objects: Array<ObjectBase>): void => {
            api.bulk(
                objects.map(obj => ({
                    endpoint: `/${props.type}/${obj.id}/poll`,
                    method: "POST",
                })),
            ).then(() => {
                getObjects();
            }, onBulkRejected);
        };

        const toggleProviders = async (objects: Array<Provider>, state: boolean): Promise<void> => {
            const ok = await ui.modal<boolean>("confirm", {
                button: {
                    label: i18n.t(`labels.providers.${state ? "enable" : "disable"}`, objects.length),
                    danger: !state,
                    primary: state,
                },
                message: i18n.t(
                    `messages.providers.${state ? "enable" : "disable"}`,
                    objects.length > 1 ? {count: objects.length} : objects[0],
                    objects.length,
                ),
            } as ModalConfirmParams);

            if (ok) {
                api.bulk(
                    objects.map(obj => ({
                        endpoint: `/providers/${obj.id}`,
                        method: "PATCH",
                        data: {
                            enabled: state,
                        },
                    })),
                ).then(() => {
                    ui.notify(
                        i18n.t(`messages.${props.type}.${state ? "enabled" : "disabled"}`, objects.length),
                        "success",
                    );

                    getObjects();
                }, onBulkRejected);
            }
        };

        const toggleSort = (key: string): void => {
            const desc = options.value.sort.startsWith("-");
            const sort = desc ? options.value.sort.substr(1) : options.value.sort;
            options.value.sort = key === sort && !desc ? `-${key}` : key;
        };

        const updateTitle = (): void => {
            ui.title(`${i18n.t(`labels.${props.type}._`, 2)} – ${i18n.t("labels.adminPanel")}`);
        };

        onBeforeMount(() => {
            const query = router.currentRoute.value.query as Record<string, string>;

            if (query.filter) {
                options.value.filter = query.filter;
                filter.value = query.filter;
            }
            if (query.kind) {
                options.value.kind = query.kind;
            }
            if (query.page) {
                options.value.page = parseInt(query.page, 10) ?? 1;
            }
            if (query.sort) {
                options.value.sort = query.sort;
            }

            watch(
                options,
                (to: Options): void => {
                    const query: Record<string, string> = {};

                    if (to.filter !== "") {
                        query.filter = to.filter;
                    }
                    if (to.kind !== "plain") {
                        query.kind = to.kind;
                    }
                    if (to.page !== 1) {
                        query.page = to.page.toString();
                    }
                    if (to.sort !== "name") {
                        query.sort = to.sort;
                    }

                    router.replace({query});

                    getObjects();
                },
                {deep: true, immediate: true},
            );
        });

        onMounted(() => updateTitle());

        watch(
            () => props.type,
            () => {
                if (!isEqual(unref(options), defaultOptions)) {
                    options.value = Object.assign({}, defaultOptions);
                } else {
                    getObjects();
                }

                updateTitle();
            },
        );

        return {
            applyFilter,
            clearSelection,
            cloneObject,
            deleteObjects,
            erred,
            filter,
            formatDate,
            getObjects,
            i18n,
            limit,
            loading,
            objects,
            options,
            pollProviders,
            selection,
            selectionEnabled,
            table,
            tabs,
            templatable,
            toggleProviders,
            toggleSort,
            total,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../mixins";

.v-content {
    @include content;

    .v-table {
        .v-flex {
            align-items: center;

            &.sort {
                cursor: pointer;
                display: inline-flex;
                height: 2rem;
            }

            .v-icon {
                margin-left: 0.35rem;

                &.disabled,
                &.linked {
                    color: var(--light-gray);
                }

                &.enabled {
                    color: var(--green);
                }

                &.error {
                    color: var(--red);
                }
            }
        }

        .v-table-row.selected .v-flex.row .v-icon {
            color: var(--toolbar-row-selected-color);
            opacity: 0.65;
        }
    }
}
</style>
