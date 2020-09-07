/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {Parser, Token, TokenType} from "../parser";

export function formatExpr(text: string, compact = false): string {
    return (compact ? formatCompact : formatDefault)(new Parser(text).tokens());
}

function formatCompact(tokens: Array<Token>): string {
    let out = "";
    let prev: TokenType | null = null;

    tokens.forEach(tok => {
        switch (tok.type) {
            case TokenType.SPACE:
            case TokenType.NEWLINE:
                return;

            case TokenType.RPAREN:
            case TokenType.RBRACE:
                if (prev === TokenType.COMMA) {
                    out = out.slice(0, -1);
                }
                out += tok.text;

                break;

            default:
                out += tok.text;
        }

        prev = tok.type;
    });

    return out;
}

function formatDefault(tokens: Array<Token>): string {
    let out = "";
    let indent = 0;
    let prev: TokenType | null = null;

    tokens.forEach(tok => {
        if (
            indent > 0 &&
            ![TokenType.SPACE, TokenType.NEWLINE, TokenType.RBRACE, TokenType.RPAREN].includes(tok.type) &&
            prev !== null &&
            [TokenType.COMMA, TokenType.LBRACE, TokenType.LPAREN].includes(prev)
        ) {
            out += "\t".repeat(indent);
        }

        switch (tok.type) {
            case TokenType.SPACE:
            case TokenType.NEWLINE:
                return;

            case TokenType.LPAREN:
            case TokenType.LBRACE:
                out += `${tok.text}\n`;
                indent++;

                break;

            case TokenType.RPAREN:
            case TokenType.RBRACE:
                if (prev !== TokenType.COMMA) {
                    out += ",\n";
                }

                indent--;
                out += `${"\t".repeat(indent) + tok.text}`;

                break;

            case TokenType.COMMA:
                out += `${tok.text}\n`;

                break;

            default:
                out += tok.text;
        }

        prev = tok.type;
    });

    return out;
}
