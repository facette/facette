/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {matcherFromParser} from "../labels/parse";
import {Parser, TokenType} from "../parser";

import {AliasExpr, Expr, MatcherExpr, ScaleExpr} from ".";

function parseAliasExpr(parser: Parser): AliasExpr {
    const expr = exprFromParser(parser);
    parser.expect(TokenType.COMMA);
    const tok = parser.expect(TokenType.STRING);

    return {type: "alias", expr, alias: tok.text} as AliasExpr;
}

function parseScaleExpr(parser: Parser): ScaleExpr {
    const expr = exprFromParser(parser);
    parser.expect(TokenType.COMMA);
    const tok = parser.expect(TokenType.NUMBER);

    return {type: "scale", expr, factor: Number(tok.text)} as ScaleExpr;
}

function exprFromParser(parser: Parser): Expr {
    const tok = parser.peek();

    switch (tok.type) {
        case TokenType.LBRACE: {
            return {type: "matcher", matcher: matcherFromParser(parser)} as MatcherExpr;
        }

        case TokenType.IDENT: {
            if (parser.peekChar() !== "(") {
                return {type: "matcher", matcher: matcherFromParser(parser)} as MatcherExpr;
            }

            // Skip both function ident and left parenthesis
            parser.next();
            parser.next();

            let expr: Expr;

            switch (tok.text) {
                case "alias":
                    expr = parseAliasExpr(parser);
                    break;

                // case "avg":
                // case "sum":
                //     expr = parseAggregateExpr(parser, tok.text);
                //     break;

                // case "sample":
                //     expr = parseSampleExpr(parser);
                //     break;

                case "scale":
                    expr = parseScaleExpr(parser);
                    break;

                default:
                    throw Error(`unknown function ${tok.text}() at ${tok.pos.line}:${tok.pos.char}`);
            }

            // Allow extraneous comma
            if (parser.peek().type === TokenType.COMMA) {
                parser.next();
            }

            parser.expect(TokenType.RPAREN);

            return expr;
        }
    }

    throw Error(`unexpected ${tok.type} at ${tok.pos.line}:${tok.pos.char}`);
}

export function parseExpr(text: string): Expr {
    const parser = new Parser(text);
    const expr = exprFromParser(parser);

    parser.expect(TokenType.EOF);

    return expr;
}
