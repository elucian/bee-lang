# Bee Specification: Graphics (13-graphics.md)

## 1. Geometric Primitives
```ebnf
point_crt   ::= "{" "x" ":" rational "," "y" ":" rational "}" ;
point_pol   ::= "{" "r" ":" rational "," "θ" ":" angle "}" ;
circle      ::= "{" "o" ":" point_crt "," "r" ":" rational "}" ;
```

## 2. Drawing API
```ebnf
draw_stmt   ::= "draw" identifier "on" identifier ";" ;
wipe_stmt   ::= "wipe" identifier ";" ;
visibility  ::= ("show" | "hide") identifier ";" ;
```

## 3. Operational Semantics
- **Coordinate Space:** Cartesian. Coordinates represent absolute units (Rational numbers).
- **Transformation:** Shapes maintain origin `o` and rotation `θ`. Rotation applies to the shape's local coordinate system.
- **Rendering:** `Canvas` object acts as the root render target. `Layer` objects enable Z-index grouping.
- **Memory:** Graphic primitives are treated as immutable value-types (copy-by-value) unless enclosed in a `[]` mutability box.
