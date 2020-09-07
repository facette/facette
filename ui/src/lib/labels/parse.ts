/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

import {Parser, TokenType} from "../parser";

import {Matcher, NameLabel, Op} from ".";

export function matcherFromParser(parser: Parser): Matcher {
    const matcher: Matcher = [];

    let tok = parser.peek();
    const isName = tok.type === TokenType.IDENT;

    if (isName) {
        matcher.push({op: Op.EQ, name: NameLabel, value: JSON.stringify(tok.text)});
        parser.next();
    }

    // Stop if not a left brace and name was at least provided
    tok = parser.peek();
    if (tok.type !== TokenType.LBRACE && isName) {
        return matcher;
    }

    tok = parser.expect(TokenType.LBRACE);

    if (parser.peek().type !== TokenType.RBRACE) {
        for (;;) {
            while ([TokenType.SPACE, TokenType.NEWLINE].includes(parser.peek().type)) {
                parser.next();
            }

            tok = parser.expect(TokenType.IDENT);
            const name = tok.text;

            tok = parser.next();
            let op: Op;

            switch (tok.type) {
                case TokenType.EQ:
                    op = Op.EQ;
                    break;

                case TokenType.NEQ:
                    op = Op.NEQ;
                    break;

                case TokenType.EQREGEXP:
                    op = Op.EQREGEXP;
                    break;

                case TokenType.NEQREGEXP:
                    op = Op.NEQREGEXP;
                    break;

                default:
                    throw Error(`expected operator but got ${tok.type} at ${tok.pos.line}:${tok.pos.char}`);
            }

            tok = parser.expect(TokenType.STRING);

            matcher.push({op, name, value: tok.text});

            // Allow extraneous comma
            if (parser.peek().type !== TokenType.COMMA) {
                break;
            }

            parser.next();

            if (parser.peek().type === TokenType.RBRACE) {
                break;
            }
        }
    }

    parser.expect(TokenType.RBRACE);

    return matcher;
}

export function parseMatcher(text: string): Matcher {
    const parser = new Parser(text);
    const matcher = matcherFromParser(parser);

    parser.expect(TokenType.EOF);

    return matcher;
}
