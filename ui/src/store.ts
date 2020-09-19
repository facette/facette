/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {RouteLocationNormalized} from "vue-router";
import {MutationPayload, MutationTree, Store, createStore} from "vuex";

import {APIError, DashboardItem, Options, TimeRange} from "types/api";
import {Modifiers} from "types/store";
import {Notification} from "types/ui";

const persistKey = "facette";

export class State {
    public apiOptions: Options = {
        connectors: [],
        driver: {
            name: "",
            version: "",
        },
    };

    public autoPropagate = true;

    public basket: Array<DashboardItem> = [];

    public error: APIError = null;

    public loading = true;

    public locale = "en";

    public modifiers: Modifiers = {
        alt: false,
        shift: false,
    };

    public routeData: Record<string, unknown> | null = null;

    public routeGuarded = false;

    public shortcuts = true;

    public pendingNotification: Notification | null = null;

    public prevRoute: RouteLocationNormalized | null = null;

    public sidebar = true;

    public theme: string | null = null;

    public timeRange: TimeRange | null = null;

    public timezoneUTC = false;
}

const state = new State();

const store = createStore({
    state,
    mutations: Object.getOwnPropertyNames(state).reduce((tree: MutationTree<State>, name: string) => {
        tree[name] = (state: State, value: unknown): void => {
            (state as any)[name] = value;
        };
        return tree;
    }, {}),
    plugins: [persist()],
});

function persist() {
    const save = (state: State): void => {
        localStorage.setItem(
            persistKey,
            JSON.stringify({
                autoPropagate: state.autoPropagate,
                basket: state.basket,
                locale: state.locale,
                pendingNotification: state.pendingNotification,
                prevRoute:
                    state.prevRoute !== null
                        ? {
                              name: state.prevRoute.name,
                              params: state.prevRoute.params,
                          }
                        : null,
                shortcuts: state.shortcuts,
                sidebar: state.sidebar,
                theme: state.theme,
                timezoneUTC: state.timezoneUTC,
            }),
        );
    };

    let saveTimeout: number | null = null;

    const saveDebounce = (mutation: MutationPayload, state: State) => {
        if (saveTimeout !== null) {
            clearTimeout(saveTimeout);
        }

        saveTimeout = setTimeout(() => save(state), 250);
    };

    return (store: Store<State>) => {
        // Restore state keys from local storage, then subscribe to mutations
        // to keep it in-sync with live changes.
        try {
            const value = localStorage.getItem(persistKey);
            if (value !== null) {
                store.replaceState(Object.assign({}, store.state, JSON.parse(value)));
            }
        } catch (err) {}

        store.subscribe(saveDebounce);

        // Ensure local storage is in-sync before unloading
        window.addEventListener("beforeunload", () => {
            if (saveTimeout !== null) {
                clearTimeout(saveTimeout);
            }

            save(store.state);
        });
    };
}

export default store;
