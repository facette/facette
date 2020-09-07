/**
 * Copyright (c) 2020, The Facette Authors
 *
 * Licensed under the terms of the BSD 3-Clause License; a copy of the license
 * is available at: https://opensource.org/licenses/BSD-3-Clause
 */

const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";

const digits = "0123456789";

const spaces = " \t";

function isDigit(c: string): boolean {
    return digits.includes(c);
}

function isIdentChar(c: string): boolean {
    return (alpha + digits + "_").includes(c);
}

export interface Token {
    type: TokenType;
    pos: Position;
    text: string;
}

export enum TokenType {
    INVALID = "invalid",
    EOF = "end of input",

    IDENT = "ident", // ident
    NUMBER = "number", // 123.45
    STRING = "string", // "abc" or 'abc'
    SPACE = "space", // " " or \t
    NEWLINE = "new line", // \n
    BADESCAPE = "bad escape", // \b

    EQ = "equal", // =
    NEQ = "not equal", // !=
    EQREGEXP = "equal pattern", // =~
    NEQREGEXP = "not equal pattern", // !~

    LBRACE = "left brace", // {
    RBRACE = "right brace", // }
    LPAREN = "left parenthesis", // (
    RPAREN = "right parenthesis", // )
    COMMA = "comma", // ,
}

export interface Position {
    line: number;
    char: number;
}

export class Parser {
    private emitSpaces: boolean;

    private last: Position | null = null;

    private lastCh: string | null = null;

    private peeked: Token | null = null;

    private pos: Position = {line: 1, char: 1};

    private text: string;

    public constructor(text: string, emitSpaces = false) {
        this.text = text;
        this.emitSpaces = emitSpaces;
    }

    public expect(type: string): Token {
        const tok = this.next();
        if (tok.type !== type) {
            throw Error(`expected ${type} but got ${tok.type} at ${tok.pos.line}:${tok.pos.char}`);
        }

        return tok;
    }

    public next(): Token {
        let tok: Token;

        if (this.peeked !== null) {
            tok = this.peeked;
            this.peeked = null;
        } else {
            tok = this.scan();
        }

        return tok;
    }

    public peek(): Token {
        if (this.peeked === null) {
            this.peeked = this.scan();
        }

        return this.peeked;
    }

    public peekChar(): string {
        const c = this.read();
        this.unread();
        return c;
    }

    public tokens(): Array<Token> {
        const tokens: Array<Token> = [];

        let tok = this.next();
        while (tok.type !== TokenType.EOF) {
            tokens.push(tok);
            tok = this.next();
        }

        return tokens;
    }

    private read(): string {
        const c: string = this.text.slice(0, 1);
        this.text = this.text.slice(1);

        this.last = Object.assign({}, this.pos);
        this.lastCh = c;

        if (c === "\n") {
            this.pos.line++;
            this.pos.char = 1;
        } else {
            this.pos.char++;
        }

        return c;
    }

    private run(set: string): string {
        let s = "";

        for (;;) {
            const c = this.read();
            if (c === "") {
                break;
            } else if (!set.includes(c)) {
                this.unread();
                break;
            }

            s += c;
        }

        return s;
    }

    private scan(): Token {
        let c = this.read();

        if (!this.emitSpaces) {
            while (c !== "" && (spaces + "\n").includes(c)) {
                c = this.read();
            }
        }

        const pos = Object.assign({}, this.pos);

        if (c === "") {
            return this.tokenAtPos(TokenType.EOF, pos);
        } else if (c === "-" || isDigit(c)) {
            this.unread();
            return this.tokenAtPos(TokenType.NUMBER, pos, this.scanNumber());
        } else if (isIdentChar(c)) {
            this.unread();
            return this.tokenAtPos(TokenType.IDENT, pos, this.scanIdent());
        } else if (c === '"' || c === "'") {
            this.unread();
            return this.tokenAtPos(TokenType.STRING, pos, this.scanString());
        } else if (spaces.includes(c)) {
            this.unread();
            return this.tokenAtPos(TokenType.SPACE, pos, this.run(spaces));
        } else if (c === "\n") {
            return this.tokenAtPos(TokenType.NEWLINE, pos, c);
        } else if (c === "\\") {
            const next = this.read();
            return this.tokenAtPos(TokenType.BADESCAPE, pos, c + next);
        } else if (c === "=") {
            const next = this.read();
            if (next === "~") {
                return this.tokenAtPos(TokenType.EQREGEXP, pos, "=~");
            }

            this.unread();

            return this.tokenAtPos(TokenType.EQ, pos, "=");
        } else if (c === "!") {
            const next = this.read();
            if (next === "=") {
                return this.tokenAtPos(TokenType.NEQ, pos, "!=");
            } else if (next === "~") {
                return this.tokenAtPos(TokenType.NEQREGEXP, pos, "!~");
            }

            this.unread();
        } else if (c === "{") {
            return this.tokenAtPos(TokenType.LBRACE, pos, c);
        } else if (c === "}") {
            return this.tokenAtPos(TokenType.RBRACE, pos, c);
        } else if (c === "(") {
            return this.tokenAtPos(TokenType.LPAREN, pos, c);
        } else if (c === ")") {
            return this.tokenAtPos(TokenType.RPAREN, pos, c);
        } else if (c === ",") {
            return this.tokenAtPos(TokenType.COMMA, pos, c);
        }

        return this.tokenAtPos(TokenType.INVALID, pos, c);
    }

    private scanIdent(): string {
        let s = "";

        for (;;) {
            const c = this.read();
            if (c !== "" && isIdentChar(c)) {
                s += c;
            } else {
                this.unread();
                break;
            }
        }

        return s;
    }

    private scanNumber(): string {
        let s = "";

        let next = this.read();
        if (next === "-") {
            s += next;
        } else {
            this.unread();
        }

        s += this.run(digits);

        next = this.read();
        if (next === ".") {
            s += next + this.run(digits);
        } else {
            this.unread();
        }

        next = this.read();
        if ("eE".includes(next)) {
            s += next;

            next = this.read();
            if ("+-".includes(next)) {
                s += next;
            } else {
                this.unread();
            }

            s += this.run(digits);
        } else {
            this.unread();
        }

        return s;
    }

    private scanString(): string {
        const quote = this.read();

        let s = "";
        let terminated = false;

        loop: for (;;) {
            const c = this.read();

            switch (c) {
                case quote:
                    terminated = true;
                    break loop;

                case "":
                    break loop;

                case "\\": {
                    const next = this.read();
                    switch (next) {
                        case "\\":
                        case '"':
                        case "'":
                            s += "\\" + next;
                            break;

                        default:
                            this.unread();
                            this.text = "\\" + this.text;

                            break loop;
                    }

                    break;
                }

                default:
                    s += c;
            }
        }

        if (terminated) {
            s += quote;
        }

        return quote + s;
    }

    private tokenAtPos(type: TokenType, pos: Position, text = ""): Token {
        return {type, pos, text};
    }

    private unread(): void {
        if (this.last !== null && this.lastCh !== null) {
            this.text = this.lastCh + this.text;
            this.pos = this.last;
            this.last = null;
            this.lastCh = null;
        }
    }
}
