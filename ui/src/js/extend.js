if (!String.prototype.startsWith) {
    String.prototype.startsWith = function(string) {
        return this.substr(0, string.length) === string;
    };
}

if (!String.prototype.endsWith) {
    String.prototype.endsWith = function(string) {
        return this.substr(-(string.length)) === string;
    };
}

if (!String.prototype.matchAll) {
    String.prototype.matchAll = function(re) {
        var matches = [],
            match,
            str,
            idx;

        // Force RegExp global flag
        if (!re.global) {
            str = re.toString();
            idx = str.lastIndexOf('/');
            re = new RegExp(str.substr(str.indexOf('/') + 1, idx - 1), str.substr(idx + 1) + 'g');
        }

        while ((match = re.exec(this))) {
            if (match.length === 0) {
                continue;
            }

            matches.push(match);
        }

        return matches;
    };
}
