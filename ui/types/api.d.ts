/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

export declare type APIError = "notFound" | "unhandled" | null;

export declare interface APIResponse<T> {
    data?: T;
    total?: number;
    error?: string;
}

export declare interface BulkRequest {
    endpoint: string;
    method: string;
    params?: Record<string, unknown>;
    data?: unknown;
}

export declare interface BulkResult {
    status: number;
    headers: Record<string, string>;
    response: APIResponse<unknown>;
}

export declare interface ListParams {
    limit?: number;
    offset?: number;
    sort?: string;
    [key: string]: unknown;
}

export declare interface Labels {
    entries(name?: boolean): Record<string, string>;
    name(): string | null;
    toString(): string;
}

export declare interface Options {
    connectors: Array<string>;
    driver: {
        name: string;
        version: string;
    };
}

export declare interface Point {
    0: number;
    1: number | null;
}

export declare interface SeriesQuery {
    from?: string;
    to?: string;
    step?: number | string;
    exprs: Array<string>;
}

export declare interface SeriesResult {
    from: string;
    to: string;
    step: number;
    series: Array<Series>;
}

export declare interface Series {
    name: string;
    points: Array<Point>;
    summary: SeriesSummary;
}

export declare type SeriesSummary = Record<string, number>;

export declare interface TestResult {
    success: boolean;
}

export declare interface TimeRange {
    from?: string;
    to?: string;
}

export declare interface Version {
    version?: string;
    branch?: string;
    revision?: string;
    compiler?: string;
    buildDate?: string;
}

// Objects
export declare interface ObjectBase {
    id: string;
    name: string;
    createdAt?: Date;
    modifiedAt?: Date;
}

export declare interface Chart extends ObjectBase {
    options?: ChartOptions;
    series?: Array<ChartSeries>;
    link?: string;
    template?: boolean;
}

export declare interface ChartOptions {
    axes?: ChartAxes;
    legend?: boolean;
    markers?: Array<Marker>;
    title?: string;
    type?: ChartType;
    variables?: Array<TemplateVariable>;
}

export declare interface ChartAxes {
    x?: ChartXAxis;
    y?: ChartYAxes;
}

export declare interface ChartXAxis {
    show?: boolean;
}

export declare interface ChartYAxes {
    center?: boolean;
    left?: ChartYAxis;
    right?: ChartYAxis;
    stack?: StackMode;
}

export declare interface ChartYAxis {
    show?: boolean;
    label?: string;
    max?: number;
    min?: number;
    unit?: Unit;
}

export declare interface Marker {
    color?: string;
    label?: string;
    value: number;
    axis?: "left" | "right";
}

export declare type StackMode = "" | "normal" | "percent";

export declare interface Unit {
    type?: UnitType;
    base?: string;
}

export declare type UnitType = "" | "binary" | "count" | "duration" | "metric";

export declare type ChartType = "area" | "bar" | "line";

export declare interface ChartSeries {
    expr: string;
    options?: ChartSeriesOptions;
}

export declare interface ChartSeriesOptions {
    color?: string;
    axis?: "left" | "right";
}

export declare interface Dashboard extends ObjectBase {
    options?: DashboardOptions;
    layout?: GridLayout;
    items?: Array<DashboardItem>;
    parent?: string;
    link?: string;
    template?: boolean;
    references?: Array<Reference>;
}

export declare interface DashboardOptions {
    title?: string;
    variables?: Array<TemplateVariable>;
}

export declare interface DashboardItem {
    type: DashboardItemType;
    layout: GridItemLayout;
    options?: Record<string, unknown>;
}

export declare type DashboardItemType = "chart" | "text";

export declare interface GridLayout {
    columns: number;
    rowHeight: number;
    rows: number;
}

export declare interface GridItemLayout {
    x: number;
    y: number;
    w: number;
    h: number;
}

export declare interface Provider extends ObjectBase {
    connector: ProviderConnector;
    filters?: Array<FilterRule>;
    pollInterval?: number;
    enabled?: boolean;
    error?: string;
}

export declare interface ProviderConnector {
    type: string;
    settings: Record<string, unknown>;
}

export declare interface FilterRule {
    action: FilterAction;
    label: string;
    pattern: string;
    into?: string;
    targets?: Record<string, string>;
}

export declare type FilterAction = "discard" | "relabel" | "rewrite" | "sieve";

export declare interface Reference {
    type: string;
    value: unknown;
}

export declare interface TemplateVariable {
    name: string;
    value?: string;
    label?: string;
    filter?: string;
    dynamic: boolean;
}
