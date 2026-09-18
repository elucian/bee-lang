// Bee syntax highlighter - single-pass tokenizer.
// Classifies tokens in one regex scan (a fast replacement for the original
// ~50 sequential .replace() passes). Keyword tables mirror
// spec/01-lexical-structure.md and spec/02-statements.md.

const APPLY_STYLE_CACHE = new Map()
const APPLY_STYLE_CACHE_LIMIT = 4096

// Style tables: [wrapper, [reserved words]]. Words are sorted longest-first
// within each group at regex build time (maximal munch).
const STYLE_TABLE = [
    [keyword,   ["use","type","object","module","class","alias","hide","from","rule","return"]],
    [imperative,["void","new","set","let","put","pop","print","read","write","apply","begin","wait","scrap"]],
    [types,     ["Ordinal","List","Array","Vector","Matrix","Set","Hash"]],
    [control,   ["start","do","done","if","else","task","with","cycle","while","for",
                 "match","when","all","one","case","trial","try","error","final","other","then","yield"]],
    [interrupt, ["assert","expect","pass","abort","exit","panic","retry","raise","resume",
                 "continue","stop","redo","next","repeat"]],
    [operator,  ["as","in","or","and","not"]],
    [builtin,   ["self","super"]],
]

// String literal alternation: single-quoted, raw backtick, double-quoted.
const STRING_ALT = [
    /'[^'\\]*(?:\\.[^'\\]*)*'/,
    /`[^`\r\n]*`/,
    /"[^"\\]*(?:\\.[^"\\]*)*"/,
].map(r => "(?:" + r.source + ")").join("|")

// One master regex: capture group 1 is the string literal; groups 2..N map to
// the keyword classes in STYLE_TABLE order. Longest words first per group.
const CODE_STYLE_REGEX = new RegExp(
    "(" + STRING_ALT + ")|" +
    STYLE_TABLE.map(function (t) {
        const words = t[1].slice().sort(function (a, b) { return b.length - a.length; })
        return "(\\b(?:" + words.join("|") + ")\\b)"
    }).join("|"),
    "g"
)

function apply_style(str) {
    if (APPLY_STYLE_CACHE.has(str)) {
        return APPLY_STYLE_CACHE.get(str)
    }

    const styled = str.replace(CODE_STYLE_REGEX, function (m, sMatch) {
        if (sMatch !== undefined) {
            return strings(sMatch)
        }
        for (let i = 0; i < STYLE_TABLE.length; i++) {
            const group = arguments[2 + i]
            if (group !== undefined) {
                return STYLE_TABLE[i][0](group)
            }
        }
        return m
    })

    if (APPLY_STYLE_CACHE.size >= APPLY_STYLE_CACHE_LIMIT) {
        APPLY_STYLE_CACHE.clear()
    }
    APPLY_STYLE_CACHE.set(str, styled)
    return styled
}

function bee_render() {
    const bee_code = document.getElementsByClassName("language-bee");
    if (typeof(bee_code) != "undefined") {
        let i = 0
        let t = ""
        let comment = ""
        let start_comments = false
        for (const e of bee_code ) {
            if (e.tagName =="CODE") {
                const lines = e.innerText.split("\n")
                // format each line
                for (let line of lines) {
                    const trimmed = line.trim()
                    //check if line is empty
                    if (i == 0 && line =="") {
                        i += 1
                        continue
                    }
                    //check if start with comments
                    if (trimmed.startsWith("+-") || start_comments) {
                        start_comments = true
                        line = blockComment(line)
                    } else if (trimmed.startsWith("--")) {
                        line = preserveIndentation(line, docComment)
                        start_comments = false
                    } else {
                        //split away end comments //
                        const separatorIndex = line.indexOf(" -- ")
                        if (separatorIndex >= 0) {
                            const originalLine = line
                            line = originalLine.slice(0, separatorIndex)
                            comment = "-- " + originalLine.slice(separatorIndex + 4)
                        } else {
                            comment = ""
                        }
                        // single pass per line: strings + keywords
                        line = apply_style(line)
                        //reattach comments
                        if (comment!="") {
                            line = line + inlineComment(comment)
                        }
                    }
                    //add new line if required
                    i += 1
                    if (i < lines.length || line!="") {
                       t += line_span(line)
                    }
                    //check if end of comments
                    if (line.indexOf("-+") > 0) {
                        start_comments = false
                    }
                }
                start_comments = false
                e.innerHTML = t;
                t = ""; i = 0
            }
            start_comments = false
        }
    } else {
      console.log("not_found")
    }
}

function line_span(str) {
    return "<span class=\"line\">"+ str + "</span>\n"
}

function docComment(str) {
    return "<span class=\"comment-doc\">" + str + "</span>"
}

function preserveIndentation(str, formatter) {
    const match = str.match(/^\s*/)
    const indent = match ? match[0] : ""
    const body = str.slice(indent.length)
    return indent + formatter(body)
}

function comments(str) {
    return "<span class=\"comment\">" + str + "</span>"
}

function blockComment(str) {
    return "<span class=\"comment-block\">" + str + "</span>"
}

function inlineComment(str) {
    return "<span class=\"comment-inline\">" + str + "</span>"
}

function keyword(str) {
    return "<span class=\"keyword\">" + str + "</span>"
}

function imperative(str) {
    return "<span class=\"impera\">" + str + "</span>"
}

function types(str) {
    return "<span class=\"type\">"+ str + "</span>"
}

function control(str) {
    return "<span class=\"control\">"+ str + "</span>"
}

function interrupt(str) {
    return "<span class=\"interrupt\">"+ str + "</span>"
}

function operator(str) {
    return "<span class=\"operator\">"+ str + "</span>"
}

function strings(str) {
    return "<span class=\"string\">"+ str + "</span>"
}

function builtin(str) {
    return "<span class=\"builtin\">"+ str + "</span>"
}
