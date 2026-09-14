const APPLY_STYLE_CACHE = new Map()
const APPLY_STYLE_CACHE_LIMIT = 4096

function apply_style(str) {
    if (APPLY_STYLE_CACHE.has(str)) {
        return APPLY_STYLE_CACHE.get(str)
    }

    //keywords without indentation

    let styled = str
    styled = styled.replace(/^use\b/,keyword("use"))
    styled = styled.replace(/^type\b/,keyword("type"))
    styled = styled.replace(/^object\b/,keyword("object"))
    styled = styled.replace(/^module\b/,keyword("module"))
    styled = styled.replace(/^class\b/,keyword("class"))
    styled = styled.replace(/^alias\b/,keyword("alias"))
    styled = styled.replace(/^hide\b/,keyword("hide"))

    //keywords with indentation
    styled = styled.replace(/\bfrom\b/,keyword("from"))
    styled = styled.replace(/\brule\b/,keyword("rule"))
    styled = styled.replace(/\breturn\b/,keyword("return"))

    //imperative keywords
    styled = styled.replace(/\bvoid\b/,imperative("void"))
    styled = styled.replace(/\bnew\b/,imperative("new"))
    styled = styled.replace(/\bset\b/,imperative("set"))
    styled = styled.replace(/\blet\b/,imperative("let"))
    styled = styled.replace(/\bput\b/,imperative("put"))    
    styled = styled.replace(/\bpop\b/,imperative("pop")) 
    styled = styled.replace(/\bprint\b/,imperative("print"))
    styled = styled.replace(/\bread\b/,imperative("read"))
    styled = styled.replace(/\bwrite\b/,imperative("write"))
    styled = styled.replace(/\bapply\b/,imperative("apply"))
    styled = styled.replace(/\bbegin\b/,imperative("begin"))
    styled = styled.replace(/\bwait\b/,imperative("wait"))
    styled = styled.replace(/\bscrap\b/,imperative("scrap"))

    //data types keywords
    styled = styled.replace(/\bOrdinal\b/,types("Ordinal"))
    styled = styled.replace(/\bList\b/,types("List"))
    styled = styled.replace(/\bArray\b/,types("Array"))
    styled = styled.replace(/\bVector\b/,types("Vector"))
    styled = styled.replace(/\bMatrix\b/,types("Matrix"))
    styled = styled.replace(/\bSet\b/,types("Set"))
    styled = styled.replace(/\bHash\b/,types("Hash"))

    // control flow keywords
    styled = styled.replace(/\bstart\b/,control("start"))
    styled = styled.replace(/\bdo\b/,control("do"))
    styled = styled.replace(/\bdone\b/,control("done"))
    styled = styled.replace(/\bif\b/,control("if"))
    styled = styled.replace(/\belse\b/,control("else"))
    styled = styled.replace(/\btask\b/,control("task"))
    styled = styled.replace(/\bwith\b/,control("with"))
    styled = styled.replace(/\bcycle\b/,control("cycle"))
    styled = styled.replace(/\bwhile\b/,control("while"))
    styled = styled.replace(/\bfor\b/,control("for"))
    styled = styled.replace(/\bmatch\b/,control("match"))
    styled = styled.replace(/\bwhen\b/,control("when"))
    styled = styled.replace(/\btrial\b/,control("trial"))
    styled = styled.replace(/\bcase\b/,control("case"))
    styled = styled.replace(/\bmiss\b/,control("miss"))
    styled = styled.replace(/\btry\b/,control("try"))
    styled = styled.replace(/\bfinal\b/,control("final"))
    styled = styled.replace(/\bother\b/,control("other"))
    styled = styled.replace(/\bthen\b/,control("then"))

    // interruption statements
    styled = styled.replace(/\bassert\b/,interrupt("assert"))
    styled = styled.replace(/\bexpect\b/,interrupt("expect"))
    styled = styled.replace(/\bpass\b/,interrupt("pass"))
    styled = styled.replace(/\babort\b/,interrupt("abort"))
    styled = styled.replace(/\bexit\b/,interrupt("exit"))
    styled = styled.replace(/\bpanic\b/,interrupt("panic"))
    styled = styled.replace(/\bfail\b/,interrupt("fail"))
    styled = styled.replace(/\bretry\b/,interrupt("retry"))
    styled = styled.replace(/\braise\b/,interrupt("raise"))
    styled = styled.replace(/\bresume\b/,interrupt("resume"))
    styled = styled.replace(/\bcontinue\b/,interrupt("continue"))
    styled = styled.replace(/\bstop\b/,interrupt("stop"))
    styled = styled.replace(/\bredo\b/,interrupt("redo"))
    styled = styled.replace(/\bnext\b/,interrupt("next"))
    styled = styled.replace(/\brepeat\b/,interrupt("repeat")) // D15: deprecated synonym of next

    //keyword operators
    styled = styled.replace(/\bas\b/,operator("as"))
    styled = styled.replace(/\bin\b/,operator("in"))
    styled = styled.replace(/\bor\b/,operator("or"))
    styled = styled.replace(/\band\b/,operator("and"))
    styled = styled.replace(/\bnot\b/,operator("not"))

    //next operator has problems, is better not to do it
    //str = str.replace(/\s\|\s/g,operator(" | "))
    //str = str.replace(/\s\&\s/g,operator(" & "))
    //str = str.replace(/\s\~\s/g,operator(" ~ "))
    //str = str.replace(/\s\+\s/g,operator(" + "))
    //str = str.replace(/\s\-\s/g,operator(" - "))
    //str = str.replace(/\s\*\s/g,operator(" * "))
    //str = str.replace(/\s\=\s/g,operator(" = "))

    //double symbols
    //str = str.replace(/=>/,operator("=>"))
    //str = str.replace(/==/,operator("=="))
    //str = str.replace(/:=/,operator(":="))
    //str = str.replace(/\+=/,operator("+="))
    //str = str.replace(/-=/,operator("-="))
    //str = str.replace(/\/=/,operator("/="))
    //str = str.replace(/\*=/,operator("*="))
    //str = str.replace(/:=/,operator(":="))
    //str = str.replace(/::/,operator("::"))
    //str = str.replace(/<</,operator("<<"))
    //str = str.replace(/>>/,operator(">>"))
    //str = str.replace(/!=/,operator("!="))
    //str = str.replace(/!≡/,operator("!≡"))
    //str = str.replace(/<:/,operator("<:"))
    //str = str.replace(/<\+/,operator("<+"))


    // many times
    //str = str.replace(/∈/g,operator("∈"))
    //str = str.replace(/∨/g,operator("∨"))
    //str = str.replace(/∧/g,operator("∧"))
    //str = str.replace(/÷/g,operator("÷"))
    //str = str.replace(/·/g,operator("·"))
    //str = str.replace(/¬/g,operator("¬"))
    //str = str.replace(/±/g,operator("±"))
    // one time
    //str = str.replace(/≡/,operator("≡"))
    //str = str.replace(/≥/,operator("≥"))
    //str = str.replace(/≤/,operator("≤"))
    //str = str.replace(/⊕/,operator("⊕"))
    //str = str.replace(/⊖/,operator("⊖"))
    //str = str.replace(/≈/,operator("≈"))
    //str = str.replace(/≠/,operator("≠"))
    //str = str.replace(/∪/,operator("∪"))
    //str = str.replace(/∩/,operator("∩"))
    //str = str.replace(/⊂/,operator("⊂"))
    //str = str.replace(/⊃/,operator("⊃"))


    // System & built-in variables
    styled = styled.replace(/\bself\b/g,builtin("self"))
    styled = styled.replace(/\bsuper\b/g,builtin("super"))

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
                        //avoid style in strings
                        if (line.indexOf('"') > 0) {
                            const parts = line.split('"')
                            line  = ""
                            let j = 0
                            for (const part of parts) {
                                if (j == 1) {
                                    line  += strings('"' + part + '"')
                                    j = 0
                                } else {
                                    line  += apply_style(part)
                                    j = 1
                                }
                            }
                        } else {
                            line  = apply_style(line)
                        }
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
