/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {computed, watch} from "vue";
import {NavigationGuardNext, RouteLocationNormalized} from "vue-router";

import {APIError, APIResponse, BulkResult} from "types/api";
import {Modifiers} from "types/store";

import i18n from "@/i18n";
import router from "@/router";
import store from "@/store";
import ui from "@/ui";

export const namePattern = "^[a-zA-Z0-9](?:[a-zA-Z0-9-_]*[a-zA-Z0-9])?$";

let guardUnwatch: (() => void) | null = null;

function onBeforeUnload(ev: Event): void {
    ev.preventDefault();
    ev.returnValue = false;
}

function guardRoute(state: boolean): void {
    store.commit("routeGuarded", state);

    window[state ? "addEventListener" : "removeEventListener"]("beforeunload", onBeforeUnload);

    if (guardUnwatch) {
        guardUnwatch();
        guardUnwatch = null;
    }
}

export default {
    applyRouteParams: (): void => {
        const route = router.currentRoute.value;

        if (route.hash) {
            route.params.section = route.hash.substr(1);
        }
    },

    beforeRoute: async (
        to: RouteLocationNormalized,
        from: RouteLocationNormalized,
        next: NavigationGuardNext,
    ): Promise<void> => {
        if (to.path === from.path || !store.state.routeGuarded) {
            next();
            return;
        }

        const ok = await ui.modal<boolean>("confirm", {
            button: {
                label: i18n.global.t("labels.leavePage"),
                danger: true,
            },
            message: i18n.global.t("messages.unsavedLost"),
        });

        if (ok) {
            store.commit("routeGuarded", false);
            window.removeEventListener("beforeunload", onBeforeUnload);
            next();
            return;
        }

        next(false);
    },

    erred: computed((): boolean => store.state.error !== null),

    error: computed((): APIError => store.state.error),

    routeGuarded: computed((): boolean => store.state.routeGuarded),

    loading: computed((): boolean => store.state.loading),

    modifiers: computed((): Modifiers => store.state.modifiers),

    onBulkRejected: (response: APIResponse<Array<BulkResult>>): void => {
        if (response.data && response.data.filter(result => result.status >= 400).length > 0) {
            ui.notify(i18n.global.t("messages.error.bulk"), "error");
        }
    },

    onFetchRejected: (response: Response): void => {
        let error: APIError = "unhandled";

        switch (response.status) {
            case 404:
                error = "notFound";
                break;
        }

        store.commit("error", error);
    },

    prevRoute: computed((): RouteLocationNormalized | null => store.state.prevRoute),

    resetError: (): void => store.commit("error", null),

    sidebar: computed((): boolean => store.state.sidebar),

    toggleSidebar: (): void => store.commit("sidebar", !store.state.sidebar),

    unwatchGuard: (): void => guardRoute(false),

    watchGuard: (...sources: Array<unknown>): void => {
        guardRoute(false);
        guardUnwatch = watch(sources, () => guardRoute(true), {deep: true});
    },
};
