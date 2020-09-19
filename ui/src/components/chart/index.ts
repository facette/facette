/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {TimeRange} from "types/api";

export interface Range {
    unit: string;
    amount: number;
    value: string;
}

export const dateFormatDisplay = "yyyy-MM-dd HH:mm:ss";

export const dateFormatFilename = "yyyyMMddHHmmss";

export const defaultTimeRange: TimeRange = {
    from: "-1h",
    to: "now",
};

export const ranges: Array<Range> = [
    {unit: "minutes", amount: 5, value: "-5m"},
    {unit: "minutes", amount: 15, value: "-15m"},
    {unit: "minutes", amount: 30, value: "-30m"},
    {unit: "hours", amount: 1, value: "-1h"},
    {unit: "hours", amount: 3, value: "-3h"},
    {unit: "hours", amount: 6, value: "-6h"},
    {unit: "hours", amount: 12, value: "-12h"},
    {unit: "days", amount: 1, value: "-1d"},
    {unit: "days", amount: 3, value: "-3d"},
    {unit: "days", amount: 7, value: "-7d"},
    {unit: "months", amount: 1, value: "-1M"},
    {unit: "months", amount: 3, value: "-3M"},
    {unit: "months", amount: 6, value: "-6M"},
    {unit: "years", amount: 1, value: "-1y"},
];
