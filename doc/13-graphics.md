# Bee Graphics

Bee provides native graphic support, focusing on dynamic 2D drawing within Cartesian spaces.

## 1. Coordinate Systems
Bee supports both radians (π) and degree notation. 
- **Minutes/Seconds:** Use Unicode prime (′) and double prime (″).

```bee
new α ∈ G := 180°; 
new β ∈ G := 0°0′0″;
```

## 2. Graphic Types
Bee manages graphics through an object hierarchy.

| Type | Signature | Description |
| :--- | :--- | :--- |
| **Canvas** | `{o ∈ P, w,h ∈ Z, m ∈ [Layer]}` | Main drawing surface |
| **Layer** | `{c ∈ B, v ∈ B, m ∈ [Shape]}` | Grouped shapes |
| **Shape** | `{o ∈ P, s ∈ ⌂, θ ∈ G}` | Transformable shape |
| **Label** | `{o ∈ P, t ∈ S, α, β ∈ G}` | Textual graphic element |

## 3. Drawing Elements
Composite graphic primitives for 2D geometry:

| Name | Signature | Meaning |
| :--- | :--- | :--- |
| **CRT** | `{x, y ∈ Q}` | Cartesian Point |
| **POL** | `{r ∈ P, θ ∈ G}` | Polar Point |
| **VEC** | `{o, p ∈ CRT}` | Vector |
| **CRC** | `{o ∈ CRT, r ∈ P}` | Circle |
| **SQR** | `{o ∈ CRT, b ∈ P, θ ∈ G}` | Square |
| **PLG** | `{v ∈ [VEC]}` | Polygon |

## 4. Drawing API
- `draw`: Add shape to a layer.
- `wipe`: Remove shapes from a layer.
- `show`/`hide`: Manage canvas visibility.


**Read next:** [System Library](/projects/bee/library/)
