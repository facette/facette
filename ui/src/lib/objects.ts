/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import cloneDeep from "lodash/cloneDeep";

import {BulkRequest, Chart, Dashboard, DashboardItem, DashboardItemType, Reference, TemplateVariable} from "types/api";

import common from "@/common";

import api from "./api";
import {parseVariables, renderTemplate} from "./template";

export function dataFromVariables(variables: Array<TemplateVariable>): Record<string, string> {
    return (
        variables.reduce((data: Record<string, string>, variable: TemplateVariable) => {
            if (!variable.dynamic) {
                data[variable.name] = variable.value as string;
            }
            return data;
        }, {}) ?? {}
    );
}

export function parseChartVariables(chart: Chart): Array<TemplateVariable> {
    let data = "";

    chart.series?.forEach(series => {
        data += `\xff${series.expr}`;
    });

    if (chart.options?.axes?.y?.left?.label) {
        data += `\xff${chart.options.axes.y.left.label}`;
    }

    if (chart.options?.axes?.y?.right?.label) {
        data += `\xff${chart.options.axes.y.right.label}`;
    }

    if (chart.options?.title) {
        data += `\xff${chart.options.title}`;
    }

    return parseVariables(data).map(name => ({name, dynamic: false}));
}

export function parseDashboardVariables(dashboard: Dashboard): Array<TemplateVariable> {
    let data = "";

    if (dashboard.options?.title) {
        data += `\xff${dashboard.options.title}`;
    }

    return parseVariables(data).map(name => ({name, dynamic: false}));
}

export function renderChart(chart: Chart, data: Record<string, string>): Chart {
    const proxy: Chart = cloneDeep(chart);

    proxy.series?.forEach(series => {
        series.expr = renderTemplate(series.expr, data);
    });

    if (proxy.options?.axes?.y?.left?.label) {
        proxy.options.axes.y.left.label = renderTemplate(proxy.options.axes.y.left.label, data);
    }

    if (proxy.options?.axes?.y?.right?.label) {
        proxy.options.axes.y.right.label = renderTemplate(proxy.options.axes.y.right.label, data);
    }

    if (proxy.options?.title) {
        proxy.options.title = renderTemplate(proxy.options.title, data);
    }

    return proxy;
}

export function renderDashboard(dashboard: Dashboard, data: Record<string, string>): Dashboard {
    const proxy: Dashboard = cloneDeep(dashboard);

    if (proxy.options?.title) {
        proxy.options.title = renderTemplate(proxy.options.title, data);
    }

    return proxy;
}

export function resolveDashboardReferences(items: Array<DashboardItem>): Promise<Array<Reference>> {
    const keys: Array<string> = [];
    const types: Array<DashboardItemType> = [];

    return api
        .bulk(
            items.reduce((req: Array<BulkRequest>, item: DashboardItem) => {
                switch (item.type) {
                    case "chart": {
                        const id = item.options?.id as string | undefined;
                        const key = `chart|${id}`;

                        if (id !== undefined && !keys.includes(key)) {
                            req.push({
                                endpoint: `/charts/${id}/resolve`,
                                method: "POST",
                            });

                            keys.push(key);
                        }

                        break;
                    }

                    default:
                        return req;
                }

                types.push(item.type);

                return req;
            }, []) ?? [],
        )
        .then(response => {
            return Promise.resolve(
                response.data?.map((result, index) => ({type: types[index], value: result.response.data})) ?? [],
            );
        });
}

export async function resolveVariables(variables: Array<TemplateVariable>): Promise<Record<string, Array<string>>> {
    const {onBulkRejected} = common;

    const req: Array<BulkRequest> = [];
    const labels: Array<string> = [];

    variables.forEach(variable => {
        if (variable.dynamic) {
            req.push({
                endpoint: `/labels/values?name=${variable.label}`,
                method: "GET",
                params: variable.filter ? {match: variable.filter} : undefined,
            });

            labels.push(variable.name);
        }
    });

    const data: Record<string, Array<string>> = {};

    if (req.length > 0) {
        await api.bulk(req).then(response => {
            response.data?.forEach((result, index) => {
                data[labels[index]] = result.response.data as Array<string>;
            });
        }, onBulkRejected);
    }

    return data;
}
