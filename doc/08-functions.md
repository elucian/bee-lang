# Bee Functions

Bee is a functional programming language. Bee enables functional programming using lambda expressions—arithmetic expressions that perform computations and return results.

## 1. Lambda Expressions
A lambda expression can accept parameters and return a result. They are stateless and lack a local scope.

**Syntax:**
```bee
new name := λ(p1 ∈ Type, ...) => (expression) ∈ ResultType;
```

**Example:**
```bee
-- Define lambda expression
new exp := λ(x, y ∈ Z) => x^y ∈ Z;

-- Use the lambda expression
new z := 2 * exp(2, 3);
print z; -- 16
```

**Properties:**
* Similar to mathematical functions.
* Can be used as callback arguments.
* Can be returned from a rule.
* Assigned to type: `L` (Lambda reference).

**Restrictions:**
* Stateless: No internal states.
* Deterministic: Side-effect free.
* Pure: Do not modify external state; cannot call `rule` entities.

## 2. Lambda Type
The type `L` is a reference used for variables, collection elements, or parameters that represent functional transformations.

**Demo: Expression Dictionary**
```bee
new gt := λ(x, y ∈ Z) => (x > y) ∈ B;
new lt := λ(x, y ∈ Z) => (x < y) ∈ B;
new eq := λ(x, y ∈ Z) => (x = y) ∈ B;

type Dic: {[A]:L} <: Map;
new dic := {'gt':gt, 'lt':lt, 'eq':eq} ∈ Dic;

print dic['gt'](3, 1); -- 1
```

## 3. Call-back functions
Lambda expressions become callbacks when passed as arguments.

**Example:**
```bee
rule foo(a1, a2 ∈ N, xp: λ(x, y ∈ N) => R) => (r ∈ R):
  let r := xp(a1, a2);
  return;

rule main:
  print foo(2, 3, (x, y) => x^y); -- 8.0
  return;
```

**Note:** A lambda cannot call a rule, ensuring it remains pure. Only a rule can invoke a lambda.

**Read next:** [Objects](/projects/bee/objects/)
