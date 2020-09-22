/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {parseMatcher} from "./parse";

export {parseMatcher};

export {matcherToString} from "./format";

export type Matcher = Array<MatcherCond>;

export interface MatcherCond {
    op: Op;
    name: string;
    value: string;
}

export enum Op {
    EQ = "=",
    NEQ = "!=",
    EQREGEXP = "=~",
    NEQREGEXP = "!~",
}

export const NameLabel = "__name__";

export const ProviderLabel = "__provider__";

export class Labels {
    private value: Record<string, string>;

    constructor(input?: string) {
        this.value = input
            ? parseMatcher(input).reduce((out: Record<string, string>, cond: MatcherCond) => {
                  out[cond.name] = JSON.parse(cond.value);
                  return out;
              }, {})
            : {};
    }

    public entries(name = true): Record<string, string> {
        if (name) {
            return Object.assign({}, this.value);
        }

        const obj = Object.assign({}, this.value);

        return Object.keys(obj).reduce((out: Record<string, string>, key: string) => {
            if (key !== NameLabel) {
                out[key] = obj[key];
            }
            return out;
        }, {});
    }

    public name(): string | null {
        return this.value[NameLabel] ?? null;
    }

    public toString(): string {
        let s = this.name() ?? "";

        const ls = this.entries(false);
        if (Object.keys(ls).length > 0) {
            s += `{${Object.keys(ls)
                .sort()
                .map(key => `${key}=${JSON.stringify(ls[key])}`)}}`;
        }

        return s;
    }
}
