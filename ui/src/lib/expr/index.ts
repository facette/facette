/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {Matcher} from "../labels";

export {formatExpr} from "./format";
export {parseExpr} from "./parse";

export interface Expr {
    type: "aggregate" | "alias" | "matcher" | "sample" | "scale";
}

export interface AggregateExpr extends Expr {
    exprs: Array<Expr>;
    op: AggregateOp;
}

export enum AggregateOp {
    AVERAGE = "avg",
    SUM = "sum",
}

export interface AliasExpr extends Expr {
    expr: Expr;
    alias: string;
}

export interface MatcherExpr extends Expr {
    matcher: Matcher;
}

export interface SampleExpr extends Expr {
    expr: Expr;
    mode: SampleMode;
}

export enum SampleMode {
    AVERAGE = "avg",
    FIRST = "first",
    LAST = "last",
    MAX = "max",
    MIN = "min",
    SUM = "sum",
}

export interface ScaleExpr extends Expr {
    expr: Expr;
    factor: number;
}
