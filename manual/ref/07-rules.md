# Bee Rules

Bee is rule oriented language. This makes Bee different and simpler than classic languages. Rules are subprograms. We design rules to split a large problem into smaller, more manageable problems. The main rule is the final orchestrator that coordinate execution of all rules in a program. We define rules statically in the code and we apply rules to resolve diverse use cases. 

## Rule Purpose

A rule can behave like a procedure, a function, a constructor, a method or a class. These are concepts used in other languages. Bee has simplified the concept under one single entity called "rule". Because a rule has it's own states, it can be instantiated and can be used to create "objects". Due to this versatility of rules, learning rule syntax is essential for understanding how to create and manage Bee code.

## Rule Anatomy

A rule is a block of code that start with keyword: "rule" and end with keyword "return". You can send arguments to a rule when you apply the rule. After one rule execute you can capture results into variables using assign statement or output parameters.

![bee rule](/projects/bee/img/bee-rule.svg)

Rule Concept

We have chosen "rule" keyword instead of "def" for naming a sub-program. Rules can have side-effects similar to procedures, methods or subroutines in other languages. Let's review what we have learned so far:

#### Review:

  * A rule start with keyword: _rule_ ;
  * A rule ends with keyword _return_ ;
  * A rule can have public/private states;
  * A rule can have parameters and results;


## Rule Features

We have learned that Bee rules can take different roles depending of how is defined. One rule have a name, parameters and result. Next we explain how to define these features to create rules and apply rules:

### Rule parameters

Parameters are special variables defined in rule signature using parenthesis. The parameter list is used to define input arguments. Each parameter is a local variable, it has a name, a type and initial value.

#### Example:

In next example we define a rule that require two string parameters. We provide arguments for each parameters by position. In this case, foo do not return results so we use "apply" keyword to execute the rule. Second parameter is input/output string so it can be modified.

. ``` \-- a rule with two parameter rule foo(name ∈ S, message ∈ [S]): let message:= "hello:" + name + ". I am Foo. Nice to meet you!"; return; \-- using apply + rule name will execute the rule rule main: new str ∈ S; apply foo("Bee", str); print str; return; ``` 

**Expected output:**

``` hello: Bee. I am Foo. Nice to meet you! ``` 

#### Notes:

  * Parameters are enumerated in a list separated by comma;
  * Optional parameters are initialized using "=";
  * Primitive types parameters receive values: _by copy_ ;
  * Composite type parameters receive value: _by share_ ;


### Rule results

A rule can have multiple results. Result variables must be declared. This is characteristic to Bee language. In other languages, you can use: "return value" but in Bee things are different. A rule has a result list similar to a parameter list.

#### Example:

In this example we have a rule that return a tuple of two values. These values can be assigned inside the rule body. If the values are not assigned the default values are used. Like parameters, the result variables can have initial values.

``` \-- rule with two results "s" and "d" \-- parameter x is mandatory y is optional rule com(x ∈ Z, y: 0 ∈ Z) => (s, d ∈ Z): let s := x + y; let d := x - y; return; rule main: \-- capture result into a single variable new r := com(3,2); -- create a list print r; -- (5,1) \-- deconstruction of result into variables: s, d new s, d := com(3,2); -- capture two values print (s, d, sep:",") ; -- 5,1 (use separator = ",") \-- ignore second result using variable "_" new a, _ := com(3); print a; -- 3 return; ``` 

#### Notes:

  * Multiple results are declared with name and can also have initial value;
  * A rule with multiple results can be called using spread operator (*)
  * You can capture results into multiple variables separated by comma;
  * You can ignore one result using anonymous variable "_";
  * Rules with multiple results can not be used in expressions;


### Variadic Rule

The last parameter in a parameter list can use prefix: "*" to receive multiple values into an array of values. This is called "varargs" parameter and is very useful way to accept multiple parameters by declaring just one.

``` \-- rule with varargs rule foo(*bar ∈ [Z]) => (x ∈ Z): new c := bar.count(); \-- precondition if (c == 0) do let x := 0; exit; done; \-- sum all parameters for ∀ i ∈ (0.!c) do let x += bar[i]; repeat; return; \-- we can call foo with variable number of arguments rule main: print foo(); -- 0 print foo(1); -- 1 print foo(1,2); -- 3 print foo(1,2,3); -- 6 print foo(1,2,3,4); -- 10 return; ``` 

### Early Termination

A rule should have a single exit point. Therefore in Bee the last statement in a rule is "return". The return is closing the rule block. However a rule can have an one or many other termination points. A function can be interrupted using keyword "exit" that terminate the rule without signaling any error. Other way to terminate a rule is to create an exception. This will be explained later.

#### Pattern:

``` \-- define a functional rule rule name(param ∈ type,...) => result ∈ type: ... exit if condition; -- early (successful) transfer ... let result := expression; -- computing the result ... return; rule main: \-- direct call and print the result print rule_name(argument,...); \-- capture rule result into a new variable: new r := rule_name(argument,...); \-- using existing variable: new n ∈ type; let n := rule_name(argument,...) return; ``` 

## Advanced Topics

Next features will be explained later in more details after we learn more elements required for advanced examples. Making examples for these features require knowledge about collections and data processing. So you can read a brief introduction now then skip ahead.

  * forward declarations
  * recursive rules
  * external rules
  * closures


### Forward declarations

Hoisting is a technique used by many modern compilers to identify declarations of members. Using this technique you can use an identifier before it is defined. In Bee there is no hoisting technique. You can not use an identifier before it is declared or loaded.

To be able to create a faster compiler we use a retro design. In Bee the main() rule is defined at the bottom of the main module. Private rules must be defined first. Public rules are defined last in the module. 

Two rules may call each other and create a cyclic interdependence. For this special case you can declare a rule "signature" before implementing it. That is called "forward declaration". Most modules do not need forward declarations.

#### Pattern:

``` \-- forward declaration pattern rule plus(a, b ∈ Z) => (r ∈ Z); -- forward declaration \-- declare the main rule rule main: \-- execute before implementation print plus(1,1); return \-- later implement the rule "plus" rule plus(a,b ∈ Z) => (r ∈ Z): let r := (a + b); return; ``` 

### Recursive Rules

A rule that call itself is so called "recursive". You should know any recursive rule can be replaced by a stack and a cycle that is much more efficient than a recursive rule.

#### Example1

Regular recursive rule can not be optimized by the compiler.

``` \-- this rule is not optimized: rule fact(n ∈ N) => (r ∈ N): when (n = 0) do let r := 1; else let r := n * fact(n-1); done; return; ``` 

#### Tail Call Optimization

TCO apply to a special case of recursion. The gist of it is, if the last thing you do in a function is call itself (e.g. it is calling itself from the "tail" position), this can be optimized by the compiler to act like iteration instead of standard recursion.

Normally during recursion, the runtime needs to keep track of all the recursive calls, so that when one returns it can resume at the previous call and so on. Keeping track of all the calls takes up space, which gets significant when the function calls itself a lot. But with TCO, it can just say "go back to the beginning, only this time change the parameter values to these new ones." It can do that because nothing after the recursive call refers to those values.

#### Example2

Compiler should be able to optimize this recursive rule.

``` \-- this rule can be optimized: rule tail(n ∈ N, acc ∈ N) => (r ∈ N): when (n = 0) do let r:= acc; else let r:= tail(n-1, acc * n); done; return; rule fact(n ∈ N) => (r ∈ N): let r := tail(n , 1); return; ``` 

#### Example3

Replacing a recursive rule with a cycle is more difficult but the rule may run faster. We encourage this design pattern and avoidance of recursive roules:

``` \-- this rule is manually optimized: rule fact(a ∈ N, b ∈ N) => (r ∈ N): while (b > 1) do let a := a * a + a; let b := b - 1; else let r := a; repeat; return; ``` 

### External rules

Will be useful to import C functions calls from Bee. These rules could be wrapped in Bee modules. We have not yet establish this is the way to go. If it is, we add a dependency toward C and I don't particularly like it. Yet if we implement this it should look maybe like this:

**Example:** This is myLib.bee file:

``` #module myLib use $bee.lib.cpp.myLib; -- load cpp library \-- define a wrapper for external "fib" rule fib(n ∈ Z) => (x ∈ Z)); let x := myLib.fib(n); return; ``` 

This is the main module:

``` \-- module main \-- load library use $bee.lib.myLib as myLib; \-- use external rule rule main: print myLib.fib(5); return; ``` 

To understand more about interacting with other languages check this article about ABI: [Application Binary Interface](https://en.wikipedia.org/wiki/Application_binary_interface)

* * *

### Closures

A closure is a special kind of rule defined inside of another rule. A closure is encapsulated in a hig order rule that can have public or private states.

#### Example:

In next example we define rule foo() that has two public states: .count and .step and a public method .next() that is a closure. You can call foo() to initialize the states.

``` \-- define foo generator rule foo(start:0 ∈ N, step:1 ∈ R): \-- create public states set .count := [start]; set .step := [step]; \-- define closure method rule .next() => (r ∈ Z): let r := foo.count + foo.step; let foo.count := r; return; return; rule main: \-- initialize the states apply foo(10, 5); \-- verify internal states expect foo.start = 10; expect foo.step = 5; \-- generate some numbers cycle: new x := 0; while x <= 30 do let x := foo.next(); print "x = " + x; done; return; ``` 

**Notes:** We use "set" to create foo() properties. These properties are alocated on the heap when foo() is first called. The heap is used to hold data for a long period of time. We box these variables using [] to be mutable, otherwise would be constant.

**Warning:** If you do not initialize the states before use, you may end-up with a run-time error when the state is used: "Uninitialized state x in line #."

#### Expected Output:

``` x = 10 x = 15 x = 20 x = 25 x = 30 ``` 

#### Notes: 

  1. Usually the attributes have default values; 
  2. There are othery ways to create generators;


**Read next:** [Objects](/projects/bee/objects/)
