/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {Matcher, NameLabel, Op} from ".";

export function matcherToString(matcher: Matcher): string {
    let out = "";
    const tmp = [...matcher];

    const idx = tmp.findIndex(a => a.op === Op.EQ && a.name === NameLabel);
    if (idx !== -1) {
        out += JSON.parse(tmp[idx].value);
        tmp.splice(idx, 1);
    }

    if (tmp.length > 0) {
        const conds = tmp.sort((a, b) => a.name.localeCompare(b.name)).map(cond => cond.name + cond.op + cond.value);
        out += `{${conds.join(",")}}`;
    }

    return out;
}
