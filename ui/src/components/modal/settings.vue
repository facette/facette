<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="settings" :title="i18n.t('labels.settings.personal')" @show="onShow">
        <template v-slot="modal">
            <v-form>
                <v-tablist :tabs="tabs" v-model:value="tab"></v-tablist>

                <template v-if="tab === 'display'">
                    <v-form>
                        <v-label>{{ i18n.t("labels.theme._") }}</v-label>

                        <div class="themes" :class="{selectable: !autoTheme}">
                            <v-theme
                                :class="{active: settings.theme === key}"
                                :key="key"
                                :name="key"
                                @click="!autoTheme && selectTheme(key)"
                                v-for="(theme, key) in themes"
                            ></v-theme>
                        </div>

                        <v-checkbox type="toggle" v-model:value="autoTheme">
                            {{ i18n.t("labels.theme.auto") }}
                        </v-checkbox>

                        <v-label>{{ i18n.t("labels.language._") }}</v-label>
                        <v-select
                            class="half"
                            :options="localeOptions"
                            :placeholder="i18n.t('labels.language.select')"
                            v-model:value="settings.locale"
                        ></v-select>

                        <v-label>{{ i18n.t("labels.timezone._") }}</v-label>
                        <v-select
                            class="half"
                            :options="timezoneOptions"
                            :placeholder="i18n.t('labels.timezone.select')"
                            v-model:value="settings.timezoneUTC"
                        ></v-select>
                    </v-form>
                </template>

                <template v-else-if="tab === 'keyboard'">
                    <v-form>
                        <v-markdown :content="i18n.t('help.keyboard.shortcuts')"></v-markdown>

                        <v-checkbox type="toggle" v-model:value="settings.shortcuts">
                            {{ i18n.t("labels.keyboard.shortcuts.enable") }}
                        </v-checkbox>
                    </v-form>
                </template>

                <template v-slot:bottom>
                    <v-button @click="modal.close()">{{ i18n.t("labels.cancel") }}</v-button>
                    <v-button primary @click="save">{{ i18n.t("labels.settings.apply") }}</v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import clone from "lodash/clone";
import {computed, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useStore} from "vuex";

import {SelectOption, Tab} from "types/ui";

import themes, {detectTheme} from "@/components/ui/themes";
import {State} from "@/store";

interface Settings {
    locale: string;
    shortcuts: boolean;
    theme: string | null;
    timezoneUTC: boolean;
}

const defaultSettings = {
    locale: "",
    shortcuts: true,
    theme: null,
    timezoneUTC: false,
};

const defaultTab = "display";

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const store = useStore<State>();

        const autoTheme = ref(false);
        const settings = ref<Settings>(clone(defaultSettings));
        const tab = ref(defaultTab);

        const tabs = ref<Array<Tab>>([
            {label: i18n.t("labels.display"), value: "display"},
            {label: i18n.t("labels.keyboard._"), value: "keyboard"},
        ]);

        const timezoneOptions = ref<Array<SelectOption>>([
            {label: i18n.t("labels.timezone.local"), value: false},
            {label: i18n.t("labels.timezone.utc"), value: true},
        ]);

        const localeOptions = computed<Array<SelectOption>>(() => {
            return Object.keys(i18n.messages.value as Record<string, unknown>)
                .sort()
                .map(name => ({label: i18n.t("name", name), value: name}));
        });

        const onShow = (): void => {
            autoTheme.value = store.state.theme === null;

            settings.value = {
                locale: store.state.locale,
                shortcuts: store.state.shortcuts,
                theme: store.state.theme ?? detectTheme(),
                timezoneUTC: store.state.timezoneUTC,
            };
        };

        const save = (): void => {
            store.commit("locale", settings.value.locale);
            store.commit("shortcuts", settings.value.shortcuts);
            store.commit("theme", autoTheme.value ? null : settings.value.theme);
            store.commit("timezoneUTC", settings.value.timezoneUTC);

            store.commit("pendingNotification", {
                text: i18n.t("messages.settings.saved", null, {locale: settings.value.locale}),
                type: "success",
            });

            location.reload();
        };

        const selectTheme = (theme: string): void => {
            settings.value.theme = theme;
        };

        watch(autoTheme, to => {
            if (to) {
                selectTheme(detectTheme());
            }
        });

        return {
            autoTheme,
            i18n,
            localeOptions,
            onShow,
            save,
            selectTheme,
            settings,
            tab,
            tabs,
            themes,
            timezoneOptions,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-modal {
    ::v-deep(.v-modal-content) {
        width: 25vw;

        .themes {
            columns: 2;
            column-gap: 1.5rem;

            .v-theme {
                border: 0.2rem solid transparent;
                border-radius: 0.2rem;
                padding: 0.5rem;

                &.active {
                    border-color: var(--accent);
                }
            }

            &.selectable .v-theme {
                cursor: pointer;
            }

            svg {
                height: auto;
                width: 100%;
            }

            & + .v-checkbox {
                margin-top: 0.75rem;
            }
        }
    }
}
</style>
