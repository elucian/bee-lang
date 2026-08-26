# Bee Specification: 2D Geometry & Graphics Engine (13-graphics.md)

## 1. Executive Graphics Engine Architecture

Bee incorporates native language-level 2D geometry primitives and drawing statements designed for engineering visualization, GUI rendering, and Cartesian space calculations:

1. **Angular Type (`G`):** Native degree (`°`), minute (`′`), and second (`″`) representation, mapped to the angular type $G \in [0^\circ \dots 360^\circ)$.
2. **2D Geometric Value Types:** Cartesian Points (`CRT`), Polar Points (`POL`), Vectors (`VEC`), Circles (`CRC`), Squares (`SQR`), and Polygons (`PLG`).
3. **Layered Graphic Hierarchy:** Rendering pipeline structured into `Canvas` root targets, `Layer` z-planes, `Shape` elements, and `Label` text markers.
4. **Declarative Drawing API:** Native imperative draw statements (`draw`, `wipe`, `show`, `hide`).

---

## 2. Angular Type (`G`) & Coordinate Systems

### 2.1 Angular Literals
Angles are expressed using mathematical degree, minute, and second symbols:
```bee
-- Angular variable declarations
new alpha ∈ G := 180°;
new beta  ∈ G := 45°30′15″;

-- Implicit conversion to Radians
new rad_val ∈ R := alpha :> R; -- Evaluates to 3.14159265 (π)
```

### 2.2 Point Representation
- **Cartesian Point (`CRT`):** `{x, y ∈ Q}` (Rational coordinate pair).
- **Polar Point (`POL`):** `{r ∈ R, θ ∈ G}` (Radius $r$ and angle $\theta$).

---

## 3. Geometric Value Primitives & Signature Catalogue

| Geometric Primitive | Type Code | Signature Layout | Description |
| :--- | :--- | :--- | :--- |
| **Cartesian Point** | `CRT` | `{x, y ∈ Q}` | Absolute Cartesian coordinate pair |
| **Polar Point** | `POL` | `{r ∈ R, θ ∈ G}` | Radial distance $r$ and angle $\theta$ |
| **Vector** | `VEC` | `{o, p ∈ CRT}` | Directed vector from origin $o$ to point $p$ |
| **Circle** | `CRC` | `{o ∈ CRT, r ∈ R}` | Center point $o$ and radius $r$ |
| **Square** | `SQR` | `{o ∈ CRT, b ∈ R, θ ∈ G}` | Anchor $o$, side length $b$, rotation $\theta$ |
| **Polygon** | `PLG` | `{v ∈ [CRT]}` | Ordered list of vertex points |

---

## 4. Graphic Hierarchy & Rendering Pipeline

Graphics are structured as an object scene graph:

```text
Canvas (Root Window Target)
└── Layer 1 (Background Z-Index 1)
│   ├── SQR (Background Grid)
│   └── CRC (Origin Marker)
└── Layer 2 (Foreground Z-Index 2)
    ├── PLG (Data Plot Line)
    └── Label (Plot Title Text)
```

### 4.1 Canvas & Layer Declarations
```bee
-- Create a 800x600 drawing canvas
new main_canvas := Canvas(o: {x: 0, y: 0}, w: 800, h: 600);

-- Create a drawing layer
new foreground_layer := Layer(id: "fg", visible: true);

-- Attach layer to canvas
apply main_canvas.add_layer(foreground_layer);
```

### 4.2 Drawing API Statements (`draw`, `wipe`, `show`, `hide`)
```bee
new circle_marker := CRC(o: {x: 10, y: 20}, r: 5);

-- Render shape onto target layer
draw circle_marker on foreground_layer;

-- Manage visibility
hide foreground_layer;
show foreground_layer;

-- Clear layer
wipe foreground_layer;
```

---

## 5. Geometric Transformations

Shapes respond to affine transformations in place:
- **`shape.rotate(θ ∈ G)`:** Rotates shape around its local origin $o$ by angle $\theta$.
- **`shape.translate(dx, dy ∈ Q)`:** Shifts origin $o$ by offsets $(dx, dy)$.
- **`shape.scale(factor ∈ R)`:** Scales shape dimensions by scale factor.

---

## 6. Formal EBNF Grammar

```ebnf
(* Angular & Point Literals *)
angle_lit         ::= integer_lit "°" [ integer_lit "′" [ integer_lit "″" ] ] ;
crt_point         ::= "{" "x" ":" expression "," "y" ":" expression "}" ;
pol_point         ::= "{" "r" ":" expression "," "θ" ":" expression "}" ;

(* Geometric Shapes *)
circle_lit        ::= "CRC" "(" "o" ":" crt_point "," "r" ":" expression ")" ;
square_lit        ::= "SQR" "(" "o" ":" crt_point "," "b" ":" expression "," "θ" ":" expression ")" ;
vector_lit        ::= "VEC" "(" "o" ":" crt_point "," "p" ":" crt_point ")" ;

(* Drawing Statements *)
draw_stmt         ::= "draw" expression "on" identifier ";" ;
wipe_stmt         ::= "wipe" identifier ";" ;
visibility_stmt   ::= ( "show" | "hide" ) identifier ";" ;
```

---

## 7. Diagnostic Error Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E1301` | `InvalidAngleFormat` | Malformed degree/minute/second angular syntax |
| `E1302` | `LayerNotFound` | Attempting to `draw` on unattached or non-existent layer |
| `E1303` | `NegativeRadius` | Circle or polar point constructed with negative radius $r$ |
| `E1304` | `PolygonVertexUnderflow` | Polygon constructed with fewer than 3 vertex points |
| `E1305` | `CanvasRenderError` | Target canvas hardware context unavailable |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized Angular type `G` (`°`, `′`, `″`), Cartesian/Polar primitives (`CRT`, `POL`, `VEC`, `CRC`, `SQR`, `PLG`), `Canvas`/`Layer`/`Shape`/`Label` scene graph, `draw`/`wipe`/`show`/`hide` statements, and diagnostic codes.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/13-graphics.md`.
