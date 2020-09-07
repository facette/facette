/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import isArray from "lodash/isArray";
import {nanoid} from "nanoid";
import {DirectiveBinding, VNode} from "vue";

import {Shortcut} from "types/ui";

import {codeMappping, macSymbols, platform} from "./vars";
import tooltip, {handle as handleTooltip} from "../tooltip";

interface ShortcutBinding {
    el: ElementWithShortcut;
    key: string;
    codes: Array<string>;
    modifiers: Record<string, boolean>;
}

type ElementWithShortcut = HTMLElement & {_binding: DirectiveBinding; _shortcut: string};

const modifierKeys = ["alt", "control", "meta", "shift"];

const shortcuts: Record<string, ShortcutBinding> = {};

let registered = false;

function keyCode(key: string): string {
    let code: string = key;

    if (code.startsWith("Digit")) {
        code = code.substr(5);
    } else if (code.startsWith("Key")) {
        code = code.substr(3);
    }

    return code.toLowerCase();
}

function handle(ev: KeyboardEvent): void {
    // Skip shortcuts handling when active element is an input, or when only a
    // modifier key has been pressed.
    if (
        document.activeElement?.matches('input, select, textarea, [contenteditable="true"]') ||
        modifierKeys.includes(ev.key.toLowerCase())
    ) {
        return;
    }

    const shortcut = Object.values(shortcuts).find((shortcut: ShortcutBinding): boolean => {
        return (
            // Handle codes having modifiers (e.g. "alt+,")
            (shortcut.codes.includes(keyCode(ev.code)) &&
                ev.altKey === shortcut.modifiers.alt &&
                ev.ctrlKey === shortcut.modifiers.control &&
                ev.metaKey === shortcut.modifiers.meta &&
                ev.shiftKey === shortcut.modifiers.shift) ||
            // Handle simple modified codes (e.g. "?" with implicit Shift modifier)
            (ev.key === shortcut.key &&
                !shortcut.modifiers.alt &&
                !ev.altKey &&
                !shortcut.modifiers.control &&
                !ev.ctrlKey &&
                !shortcut.modifiers.meta &&
                !ev.metaKey &&
                !shortcut.modifiers.shift &&
                ev.shiftKey)
        );
    });

    if (shortcut) {
        ev.preventDefault();
        shortcut.el.dispatchEvent(new CustomEvent("shortcut"));
    }
}

export default {
    beforeMount(el: ElementWithShortcut, binding: DirectiveBinding<Shortcut>, vnode: VNode): void {
        const ui = binding.instance?.$ui ?? null;
        if (ui === null || !ui.state.shortcuts.enabled) {
            return;
        }

        if (!registered) {
            document.addEventListener("keydown", handle);
            registered = true;
        }

        el._binding = {
            instance: binding.instance,
            value: binding.value.tooltipHelp ?? true ? binding.value.help : null,
            oldValue: null,
            modifiers: {},
            dir: binding.dir,
        };
        el._shortcut = nanoid(8);
        el.dataset.vShortcut = binding.value.keys;

        const keys = splitValue(binding.value.keys);

        const filteredKeys = keys.filter(key => !modifierKeys.includes(key));
        if (filteredKeys.length !== 1) {
            throw Error("shortcut should contain a single base key");
        }

        ui.state.shortcuts.entries[el._shortcut] = {
            keys: binding.value.keys,
            help: binding.value.help,
        };

        shortcuts[el._shortcut] = {
            el,
            key: filteredKeys[0],
            codes: codeMappping[filteredKeys[0]] ?? [filteredKeys[0]],
            modifiers: modifierKeys.reduce((out: Record<string, boolean>, modifier: string) => {
                out[modifier] = keys.includes(modifier);
                return out;
            }, {}),
        };

        // Only add event listener if there is no tooltip directive already
        // attached to the node
        if (el.classList.contains("v-button") && vnode.dirs?.findIndex(dir => dir.dir === tooltip) === -1) {
            el.addEventListener("mouseenter", handleTooltip.bind(el));
            el.addEventListener("mouseleave", handleTooltip.bind(el));
            el.addEventListener("mouseup", handleTooltip.bind(el));
        }
    },

    beforeUnmount(el: ElementWithShortcut, binding: DirectiveBinding<Shortcut>, vnode: VNode): void {
        if (shortcuts[el._shortcut]) {
            delete binding.instance?.$ui.state.shortcuts.entries[el._shortcut];
            delete shortcuts[el._shortcut];

            // Stop listening for shortcuts if none registered
            if (Object.keys(shortcuts).length === 0) {
                document.removeEventListener("keydown", handle);
                registered = false;
            }
        }

        if (el.classList.contains("v-button") && vnode.dirs?.findIndex(dir => dir.dir === tooltip) === -1) {
            el.removeEventListener("mouseenter", handleTooltip.bind(el));
            el.removeEventListener("mouseleave", handleTooltip.bind(el));
            el.removeEventListener("mouseup", handleTooltip.bind(el));
        }
    },
};

function splitValue(value: string): Array<string> {
    return value === "+" ? ["+"] : value.split("+");
}

export function shortcutLabel(value: string | Array<string>): string {
    if (isArray(value)) {
        value = value[0];
    } else if (typeof value !== "string") {
        throw Error(`invalid shortcut value: "${value}"`);
    }

    return splitValue(value)
        .map(key => {
            switch (platform) {
                case "mac":
                    if (macSymbols[key]) {
                        return macSymbols[key] + " ";
                    }
                    break;

                case "win":
                    if (key === "meta") {
                        return "Win";
                    }
            }

            return key.charAt(0).toUpperCase() + key.substr(1);
        })
        .join(platform === "mac" ? "" : " + ")
        .trimEnd();
}
