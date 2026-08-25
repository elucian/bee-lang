# Bee Syntax

Bee is not a curly bracket language. Its syntax is inspired from Ada, Ruby, Fortran and Julia. We have created an imperative programming language with original front-end.

## Operators

Bee operators are ASCII or Unicode symbols. One operator can be created using one or two characters. This is why Bee language is considered experimental, esoteric & disruptive.

## Delimiters

| Symbol  | Description                                         |
|---------|-----------------------------------------------------|
| +-...-+ | Multi-line boxed comments                           |
| #(....) | String interpolation (placeholder) for operator "?" |
| (_,_,_) | Expression                                          | List literal   |
| [_,_,_] | Index                                               | Array literals | Parameterize types |
| {_,_,_} | Enumeration type                                    | Set of values  | Hash map           |


## Strings

| symbol | description                                      |
|--------|--------------------------------------------------|
| `x`    | Back quoted string: regular expression.          |
| 'x'    | Single quoted string literal or ASCII code point |
| "y"    | Double quoted string literal or UTF32 code point |


## Single Symbols

| symbol | description                                                                    |
|--------|--------------------------------------------------------------------------------|
| !      | Negation symbol for relations                                                  | Excluded from domain                         |
| ?      | Template modifier. Associated with string templates                            |
| *      | String replication                                                             | Varargs prefix                               | Spread operator | Many something          |
| @      | Domain name                                                                    | Example @sagecode.org                        |
| $      | System sigil                                                                   | Las element in collection                    |
| &      | String concatenation                                                           | number concatenation                         |
| #      | Title                                                                          | String interpolation                         |
| ∈      | Define variable/constant/result/parameter type                                 |
| _      | Anonymous variable                                                             | Constant value = one space (_ = ' ')         |
| +      | Maximum upper limit for a domain                                               | Unicode notation U+                          |
| -      | Minimum lower limit in a domain                                                | Unicode notation U-                          |
| :      | Start a block or define something                                              |
| :      | Pair-up key-value in: objects, rule parameters, rule arguments, hash-map pairs |
| ;      | End of statement                                                               | Statement separator                          |
| .      | Decimals for real numbers                                                      | Path string concatenation                    |
| .      | Membership dot notation                                                        | Prefix for public member/attribute           | ,               | Enumeration of elements | expressions |
|        |                                                                                | Declarative collection builder: set := { x*2 | x ∈ (0..3)}     |
| \      | Escape character ( \n := New Line), ( \" = Double Quotes)                      |


## Numeric operators

Listed in the order of precedence top down.

| symbol | description                                     |
|--------|-------------------------------------------------|
| -      | Change sign, replace "y = -x" with "y = -1*x"   |
| /      | Rational number division                        |
| ^      | Power symbol used with fractions or expressions |
| √      | Radical: x√n is equivalent to x^(1/n)           |
| *      | Multiplication alternative                      |
| \      | Rational number division                        |
| /      | Real number division                            |
| ×      | Array multiplication                            | Matrix multiplication |
| %      | Modulo operator 5 % 2 = 2                       |
| +      | Numeric addition                                | List append           | Matrix addition |
| -      | Numeric subtraction                             | Collection difference |
| ±      | Numeric tolerance (use with ≈)                  |


## Double Symbols

Double symbols is a group of two ASCII symbols considered as one. Some of these symbols have an Unicode equivalent, some do not. When available, Unicode equivalent is preferred choice.

| symbol | description                                                    |
|--------|----------------------------------------------------------------|
| \--    | Single line, end of line comments                              |
| ..     | Define range/domain/slice (n..m)                               | [n..m]                            |
| .!     | Define range/domain with excluded limit (n.!m)                 | [n.!m]                            |
| !.     | Define range/domain with excluded limit (n!.m)                 | [n.!m]                            |
| !!     | Define range/domain with excluded limits: (n!!m)               | [n.!m]                            |
| -.     | Minus infinite domain: instead of [-∞..0] write: [-..0]        |
| .+     | Plus infinite domain: instead of [0..+∞] write: [0..+]         |
| =>     | Define: rule expression                                        | rule result                       |
| <-     | Define and generate values in a loop from range or set         |
| <:     | Define subset from set                                         | Specify super-type for a new type |
| :>     | Data cast pipeline operator / Type conversion                  |
| <<     | Shift values of collection to right by removing first elements |
| >>     | Shift values of collection to left by removing first elements  |
| ::     | Deep copy                                                      | Clone operator                    |
| ++     | Extend an array with one or more elements                      |
| -=     | Find and delete one element, from a collection                 |
| +=     | Append an element in a set or a map but not in a list          | +>                                | Append element to beginning of a list |
| <+     | Append element to end of a list                                |
| ==     | Relation operator for identical (the same)                     |
| !=     | Relation operator not identical (not the same)                 |
| ~=     | Relation operator: regular expression match                    |
| >=     | Relation operator: greater then or equal to                    |
| <=     | Relation operator: less then or equal to                       |


## Modifiers

Each modifier is created with pattern "x=" where x is a single symbol:

| symbol | meaning                 |
|--------|-------------------------|
| :=     | Modify                  | (value | reference) |
| +=     | Increment value         |
| -=     | Decrement value         |
| *=     | Multiplication modifier |
| /=     | Real division modifier  |
| ^=     | Power modifier          |
| √=     | Radical modifier        |
| %=     | Modulo modifier         |


## Relation Operators

Relation operators are used to compare expressions.

| symbol | meaning                                                       |
|--------|---------------------------------------------------------------|
| ∈      | check if element belong to collection                         |
| =      | equal { compare values or attributes}                         |
| ≠      | different { compare values or attributes}                     |
| ≡      | equivalent                                                    | { compare values / convert type } |
| ≈      | approximating equal numbers, used with ± like: (x ≈ 4 ± 0.25) |
| >      | value is greater than: (2 > 1)                                |
| <      | value is less than: (1 < 2)                                   |
| ≥      | greater than or equal to                                      |
| ≤      | less than or equal to                                         |
| ÷      | Exact divisor: 3 &division 15 ≡ True                          |


**negation:**

Operator: "!" can be used in combination with other operators:

``` x != y; -- equivalent to: ¬(x = y) x !≡ y; -- equivalent to: ¬(x ≡ y) x !∈ y; -- equivalent to: ¬(x ∈ y) x !≈ y; -- equivalent to: ¬(x ≈ y) x !~ y; -- equivalent to: ¬(x ~ y) ``` 

## Collection operators

| symbol | result  | meaning                               |
|--------|---------|---------------------------------------|
| ∩      | Set     | Intersection between two collections  |
| ∪      | Set     | Union between two collections         |
| ⊂      | Logic   | Set is included in superset: "⊂"      |
| ⊃      | Logic   | Set contain subset: "⊃"               |
| Δ      | Set     | Set symmetric difference              |
| +      | String  | Concatenation between two strings     |
| +      | List    | Concatenation between two lists       |
| +      | Array   | Concatenation between two arrays      |
| ∀      | Element | All: used in collection qualification |
| ∃      | Logic   | One: used in collection qualification |


## Logic Operators

Bee is using enumeration symbols: True = 1 and False = 0

| symbol | meaning | notes             |
|--------|---------|-------------------|
| ¬      | NOT     | unary operator    |
| ∧      | AND     | shortcut operator |
| ∨      | OR      | shortcut operator |
| ⊕      | XOR     | exclusive OR      |
| ↓      | NOR     | p ↓ q = ¬ (p ∨ q) |
| ↑      | NAND    | p ↑ q = ¬ (p ∧ q) |


#### The table of truth

| p   | q  | ¬ p | p ⊕ q | p ∧ q | p ∨ q |
|-----|----|-----|-------|-------|-------|
| 1   | 1  | 0   | 0     | 1     | 1     |
| 1   | 0  | 0   | 1     | 0     | 1     |
| 0   | 1  | 1   | 1     | 0     | 1     |
| 0   | 0  | 1   | 0     | 0     | 0     |


## Bitwise operators

Bitwise operators are processing numbers not Boolean values.

| symbol | meaning    | notes                         |
|--------|------------|-------------------------------|
| «      | bit SHIFTL | shift bits to left            |
| »      | bit SHIFTR | shift bits to right           |
| ~      | bit NOT    | negate all bits               |
| &      | bit AND    | execute AND between each bits |
|        |            | bit OR                        | execute OR between each bits |
| ⊕      | bit XOR    | execute XOR between each bits |


#### Arity = 1

| a    | ~ a  | a « 1 | a » 2 |
|------|------|-------|-------|
| 0000 | 1111 | 0000  | 0000  |
| 1111 | 0000 | 1110  | 0011  |
| 0111 | 1000 | 1110  | 0001  |
| 0110 | 1001 | 1100  | 0001  |


#### Arity = 2

| a   | b  | a & b | a  | b  | a ⊕ b |
|-----|----|-------|----|----|
| 00  | 00 | 00    | 00 | 00 |
| 01  | 00 | 00    | 01 | 01 |
| 11  | 01 | 01    | 11 | 10 |
| 10  | 11 | 10    | 11 | 01 |


## String operators

| Symbol | Description                                                  |
|--------|--------------------------------------------------------------|
| *      | string pattern repetition (right operator must be numeric)   |
| /      | concatenate url or path using / not depending on OS          |
| +      | concatenate two strings as they are preserving trial spaces. |
| -      | concatenate two strings and trim spaces to a single space.   |
| .      | concatenate strings with "/" on Linux or "\" on Windows.     |
| ?      | string format operator, replace "#" with number.             |


* * *

**Read next:** [Structure](/projects/bee/structure/)
