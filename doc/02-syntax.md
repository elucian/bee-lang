##### Lab Topics

* * *



# Bee Syntax

Bee syntax is inspired from Ada, Ruby, Fortran and Julia. We have created an imperative rule based programming language, not a curly bracket language. Bee is designed to be a compiler, syntax is tailored for faster efficient compilation.

#### Page bookmarks:

Next we enumerate the fundamental concepts to grasp Bee syntax. After this overview we will dive deeper inti details with examples. On this documentation we use long pages you can scroll with redundancy for better assimilation of concepts and aspects. We have modernized the aspect of HTML we use a sidebar for bookmarks that can be collapsed and expanded. This will help you progress faster and continue study from where you left off.

#### Syntax Legend

Examples and patterns illustrate language statements. Not all snippets are fully compilable or executable; their purpose is to demonstrate core concepts. Executable test cases are specific to each compiler implementation and serve as functional reference examples.

  * Suggestive descriptors represent language elements.
  * `...` denotes repetitive symbol sequences.
  * Notes and comments define semantic behavior.
  * Optional keywords are enclosed in square brackets `[]`.


## Comments

Comments are very important part of Bee code. We have multiple conventions for making good comments for any project. Bee comments are tailored by architectural principle: "if there are no comments in the code the code is wrong" Comments can be used for a document generator, that describe code API for libraries to be reused. 

#### Example:

In next example we are using various comments into a demo program.

``` #!/bin/bee +------------------------------------------------------------------ | At the beginning of program you can have several comments, | | to explain how the program works. This notation is preferred. | +-----------------------------------------------------------------+ rule main: ; -- this empty statement does nothing \-- this is a single line comment print ("end of line comments", -- first argument "can be used to explain", -- second argument "diverse arguments" -- third argument ); return; ``` 

#### Single line comments

You can use comments starting with: "--", these comments can be at start of new line, with indentation or can be used at end of line before new line of code: (EOL). Everything after -- to end of line is ignored by compiler and considered one single space.

  * notice one line may be or not a full statement. The end of statement is not (EOL) but ";",
  * you can use "-- " in the middle of an expression, if expression is on multiple lines,
  * you can have multiple statements separated by ";" in a line but only one comment before (EOL).


#### Block comment

Bee has a specific notation for block comments not used in any other language so far. It is a multi-line comment starting with "+-" and end with "-+". The upper right corner is missing in a box comment. I guess you will notice this defect later.

#### Notes:

  * Bee comments are inspired from Ada language and PL/SQL
  * Bee comments are designed for better syntax coloring


## Keywords

Bee is an expressive, verbose language. It's core has about 72 reserved keywords so far:

| begin | alias  | and      | apply | abort   |
|-------|--------|----------|-------|---------|
| other | case   | continue | done  | default |
| if    | is     | do       | else  | exit    |
| fail  | final  | miss     | panic | like    |
| load  | next   | job      | match | over    |
| print | pass   | void     | rule  | return  |
| fail  | retry  | none     | scrap | type    |
| read  | trial  | stop     | yield | xor     |
| write | wait   | when     | or    | with    |
| hide  | new    | cycle    | let   | set     |
| while | for    | resume   | put   | pop     |
| raise | not    | as       | in    | start   |
| try   | expect |          |       |         |


#### Notes:

  * You can not use these keywords as identifiers;
  * Some of these keywords are reserved but not implemented;
  * New keywords are going to be created for new features;


### Semantic keywords

| Keyword | Purpose                                      |
|---------|----------------------------------------------|
| if      | conditional executor for one statement block |
| is      | query element or variable data type          |
| as      | create alias for used modules                |
| or      | alternative for ladder decision              |
| in      | alternative for belong operation             |
| and     | alternative for cascade decision             |
| xor     | alternative for logic operation              |
| not     | alternative for logic operation              |


## Statements

Statements can start with imperative keyword or a declarative keyword:

#### Examples:

| set   | create a constant                         |
|-------|-------------------------------------------|
| new   | create a variable                         |
| let   | modify a variable                         |
| type  | create a data type                        |
| read  | accept input from console into a variable |
| write | register in console cash a string         |
| print | output to console with end of new line    |


#### Notes:

  * One statement is usually indented 2 space,
  * One statement is usually described in a single line,
  * Multiple statements on a single line are separated with ";",
  * One expression in a statement can extend on multiple lines.


### Code blocks

Statements can be contained in blocks of code.

| Keyword |  Block description                |
|---------|-----------------------------------|
| start   | start local scope for do block    |
| with    | qualifier suppression block       |
| if      | first block in decision statement |
| cycle   | repetitive or iterative blocks    |
| match   | multi-path value selector block   |
| trial   | exception handler block           |


#### Notes:

  * Block ending keyword can be one of: { done, cycle, return },
  * Statements in nested blocks are using indentation.


### Definition statements

Next statements are used to declare new elements in a module.

| Keyword | Purpose                                               |
|---------|-------------------------------------------------------|
| use     | Load module or module                                 |
| alias   | Eliminate scope qualifier                             |
| hide    | Hiding public members from a loaded module            |
| rule    | Create a new _business rule_ or _prototype_           |
| return  | End _rule_ declaration and transfer control to caller |


### Execution statements

Next keywords are simple statements. These represents actions called _imperative statements_.

| Keyword | Purpose                                                     |
|---------|-------------------------------------------------------------|
| apply   | Execute a _rule_ and ignore the _result_ if there is one    |
| begin   | Commence execution of a coroutine                           |
| wait    | Suspend current thread execution for a number of seconds    |
| read    | Flush the console buffer and accept user input from console |
| write   | Add something to console buffer but no new line             |
| print   | Output expression result, variable or constant to console   |
| let     | Mutate variable value using an expression                   |
| scrap   | Remove one element from its collection                      |


## Control statements

Control statements are used to create local blocks of code that resolve a small task synchronously. After task is finished the control is returned to the main thread.

| Keyword | Purpose                                                |
|---------|--------------------------------------------------------|
| start   | Create non repetitive local scope                      |
| if      | Start a conditional branch                             |
| else    | Start an alternative branch                            |
| do      | Start a block of code                                  |
| cycle   | Create repetitive local scope                          |
| for     | Create finite iterative block                          |
| while   | Create conditional repetitive block                    |
| match   | Value multi-path search selector                       |
| when    | Create node for match statement                        |
| other   | Default branch for match statement                     |
| trial   | Start declaration region for a protected block of code |
| try     | Begin the executable region in a trial statement       |
| case    | Associated with trial to resolve specific errors       |
| miss    | Default trial block, executed when there is no case    |
| final   | Associated with trial to finalize the trial block      |


## Transfer statements

These statements execute a jump or make an interruption of current thread.

| Keyword | Purpose                                                                      |
|---------|------------------------------------------------------------------------------|
| panic   | Create unrecoverable error code and stop current program                     |
| over    | Silent termination of program. No error is raised in this case.              |
| exit    | Silently stop execution of current rule and return to the caller             |
| yield   | Suspend one coroutine and give control to another routine                    |
| rest    | Suspend a routine and wait for all threads created by the routine to finish  |
| stop    | Interrupt execution for current cycle and continue after the cycle,          |
| redo    | Continue current cycle from the beginning making a shortcut,                 |
| next    | Continue current iteration from the beginning making a shortcut,             |
| abort   | stop early a trial block                                                     |
| fail    | Create error message and continue with next step                             |
| pass    | Skip the rest and continue with next step                                    |
| expect  | Does nothing if condition is true, otherwise create an $unexpected exception |
| raise   | Interrupt a try job or trial and issue an error                              |
| retry   | Repeat a trial block from the beginning                                      |
| resume  | Mark error as handled and continue trial                                     |
| done    | end a block statement                                                        |
| repeat  | end a repetitive block                                                       |


## Declarations

In Bee, all variables must be declared using an imperative statement. Variables can be dynamic or static and can have a data type. Data type can be custom or pre-defined. Next keywords are relevant for this topic.

| type | declare custom data type   |
|------|----------------------------|
| new  | declare a dynamic variable |
| set  | declare a static variable  |


### Identifiers:

Bee identifiers can start with dollar ($), dot (.), underscore (_), Latin, Greek, Cyrillic. An identifier can contain numbers but can not start with a number.

#### Unicode letters:

In mathematics is very popular notation for angles to use Greek letters. We support in Bee a limited number of Greek an Cyrillic letters for identifiers:

``` Σ Π Δ Ξ Γ Ψ Ω ζ α β ɣ λ π μ φ ε δ η σ ω Б Г Д Ж И Л Ф Ц Ч Ш Э Я ``` 

### Subscript:

You can use a limited number of letters and numbers available in Unicode as subscript to make identifier names. You can not start an identifier with one of these symbols and you can't add other symbols that are not subscript after a subscript:

``` x₀ x₁ x₂ x₃ x₄ x₅ x₆ x₇ x₈ x₉ x₁₀ yₐ yₑ yₕ yᵢ yⱼ yₖ yₗ yₘ yₙ yₒ yₚ yᵣ yₛ yₜ yᵤ yᵥ yₓ ``` 

### Superscript:

Bee has support for exponent using superscript. You can make any integer exponent including negative numbers but you can not use dot or fraction in the exponent.

``` x⁺ x⁻ x¹ x² x³ x⁴ x⁵ x⁶ x⁷ x⁸ x⁹ x¹⁰ ``` 

**Note:** Symbol (^) is exponent operator and is not required when you use superscript simple exponent. You can use it with complex expressions, constants or rational numbers to resolve the complex cases. Complex exponent expression must use x^() pattern, the parenthesis are mandatory only for expressions. Variables or constant literals do not need () for example x^y is valid notation.

## Expressions

Expressions are created using identifiers, operators, rules and constant literals. Expressions can be anonymous or can be assigned to identifiers to create lambda expressions.

expressions ...

  * can use () to establish order of operations,
  * can be enumerated using comma separator "," in a list,
  * can be combined to create more complex expressions,


#### Examples

``` \-- expressions print 10 print 10 + 10 + 15 print "this is a test" \-- complex expressions print (10 > 5) ∨ (2 < 3) print -b + sqr(b² - 4·a·b)/(2·a) \-- enumeration of expressions print (1,2,3) print (1,',',2,',',3) ``` 

### Exponent

Identifiers that start with a lowercase Latin letter can be used as exponent. The superscript variable can start with a letter and can also use numbers. Exponent superscript letters are mapped to regular letters to represent variables or constant. For example yᵃ is equivalent with y^a. Letter a or ᵃ, represent same entity.

``` yᵃ yᵇ yᶜ yᵈ yᵉ yᶠ yᵍ yʰ yⁱ yʲ yᵏ yᶩ yᵐ yⁿ yᵒ yᵖ yʳ yˢ yᵗ yᵘ yᵛ yʷ yˣ yʸ yᶻ ``` 

**Note:** Because Unicode is not perfect, Uppercase single letter exponent is not supported. If you have this case, you must use "^" symbol to realize the exponent. The initial design was to have this support but we changed the specification in 2026. We no longer support the following examples:

``` zᴬ zᴮ zᴰ zᴱ zᴲ zᴳ zᵸ zᴵ zᴶ zᴷ zᴸ zᴹ zᴺ zᴻ zᴼ zᴾ zᴿ zᵀ zᵁ zᵂ ``` 

### Conditional Execution

A condition is a logic expression used to control statement execution. For this we use {"if", "else"} keywords at end of statements.

``` \-- conditional statement execution statement if condition; ``` 

**Note:** Previous statement is executed only if the condition is True.

``` \-- alternative statement expect condition else statement; \-- alternative expression expect condition else expression; ``` 

**Note:** Previous statement is executed only if the condition is False.

**restrictions:**

  1. Can not use "if" with set statement;
  2. Can not use "if" with new statement;
  3. Can not use "if" after done;


#### Example:

``` rule main: \-- generate a random number new a := random(Z); \-- conditional execution new b := a; let b := -a if a < 0; \-- print result print "|b| = ", a; return; ``` 

#### Division Operation

Bee has support for fractions. Bee is using regular slash "/" for all fractions. You can use superscript for left and subscript for right: These two are equivalent (1/2 = ¹/₂). Unfortunately we can not support fractional power due to lower readability.

``` ¹/₂ ¹/₃ ¹/₄ ¹/₅ ¹/₆ ¹/₇ ¹/₈ ¹/₉ ¹/₁₀ ¹/₁₀₀ x⁻¹ = 1/x, x⁻² = 1/x², x⁻³ = 1/x³ ... ``` 

Power operations have priority but we have support only for (+, -) no other operations are possible in exponent. In next expressions, (n-1) is done first before making the power operation.

``` \-- equivalent notation xⁿ⁻¹ = x^(n-1) xˣ⁺¹ = x^(n+1) \-- equivalent notation x^(¹/₂) = √2(x) x^(¹/₃) = √3(x) ```   

**Note:** In expressions above () symbols are mandatory. The compiler will detect missing parenthesis and will ask for it. This will improve code readability and eliminate confusions.

### Pattern Matching

Instead of ternary operator we use conditional expressions. Conditional expressions enable many choices unlike ternary operator that enable only 2 choices. Conditional expressions are also known as pattern matching expressions.

#### Syntax:

``` rule main: \-- define a local variable new var ∈ type; \-- single condition matching let var := (xp1 if cnd1 else xp); \-- multiple matching with default value let var := (xp1 if cnd1, xp2 if cnd2,..., xp); \-- alternative code alignment let var := ( xp1 if cnd1 else xp2 if con2 else xp3 if cnd3 else xp ); return; ``` 

#### Example:

``` rule main: new x := '0'; -- symbol write "x:" read x; new kind := ("digit" if x ∈ ['0'..'9'] else "letter" if x ∈ ['a'..'z'] else "unknown"); print ("x is " + kind); -- expect: "x is digit" return; ``` 

### Errors

An error is when program enter a difficult state that is confusing. An error can be declared by the user or by the system. Bee has predefined type: Error that can be used to declare your own kind of errors. In other languages we use therm Exception, that is synonym to Error.

#### Internal Definition:

Parts of Bee compiler will be created using Bee language. Here is the definition of global variable $error, that is available for introspection after you call a rule.

``` \-- global error type define Error: {code ∈ Z, message ∈ S, line ∈ Z} <: Object; \-- global system error new $error ∈ Error; ``` 

You can define errors with code > 200 and raise error with 3 statements:

  * fail :raise error if condition is True
  * pass :raise error if condition is False


#### Pattern:

``` new my_error := {200,"message"} ∈ Error; fail my_error if condition; pass if condition; ``` 

String interpolation "?" can be used to customize the error messages:

#### Example:

``` rule main: new flag ∈ B; read (flag, "enter flag (0/1):"); new my_error: {201,"error:#(s)"} ∈ Error; fail (my_error ? "test") if flag; return; ``` 

#### Output:

``` error:"test" ``` 

#### Notes:

  * Keyword _fail_ will modify create an error message;
  * Keyword _pass_ is opposite of _fail_
  * Keyword _over_ will liberate the resources and terminate the program;
  * Error code < 200 are system reserved error codes;
  * Error code ≤ -1 are unrecoverable errors created with _panic_ ;


#### Unrecoverable:

Next we create unrecoverable error. In this case the program crash and exit. The operating system receive a number that signal the error code:

``` panic -1; -- end program immediately panic 2; -- end program and error code = 2 ``` 

**Read next:** [Operators](/projects/bee/structure/)

* * *

© 2026 Sage-Code Laboratory

☰
