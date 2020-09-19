/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {Unit} from "types/api";

import {
    BinaryUnit,
    CountUnit,
    DurationUnit,
    MetricUnit,
    formatBinary,
    formatCount,
    formatDuration,
    formatMetric,
    formatNumber,
} from "@/lib/format";

export function formatValue(input: number | null, unit?: Unit, decimals = 2): string {
    if (input === null) {
        return "null";
    }

    if (unit !== undefined) {
        switch (unit.type) {
            case "binary": {
                return formatBinary(input, decimals, unit.base as BinaryUnit);
            }

            case "count": {
                return formatCount(input, decimals, unit.base as CountUnit);
            }

            case "duration": {
                return formatDuration(input, decimals, unit.base as DurationUnit);
            }

            case "metric": {
                return formatMetric(input, decimals, unit.base as MetricUnit);
            }
        }
    }

    return formatNumber(input, decimals);
}
