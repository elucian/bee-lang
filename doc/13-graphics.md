# Bee Graphics Specification

Bee provides native graphic support, focusing on dynamic 2D drawing within Cartesian spaces. 

## Coordinate Systems

Bee supports radians (π) and decimal degree notation.

| Symbol | Degree |
|--------|--------|
| 0      | 0°0′0″ |
| π/4    | 45°    |
| π/2    | 90°    |
| π      | 180°   |
| 2π     | 360°   |


**Minutes and Seconds:** Bee uses Unicode symbols prime (′) for minutes and double prime (″) for seconds of arc.

``` 
new α ∈ G := 180°; new β ∈ G := 0°0′0″; 
``` 

## Graphic Types

| Type   | Signature                     | Description                    |
|--------|-------------------------------|--------------------------------|
| Canvas | {o ∈ P, w,h ∈ Z, m ∈ [Layer]} | Canvas with points and shapes  |
| Layer  | {c ∈ B, v ∈ B, m ∈ [Shape]}   | Layer with color and shape set |
| Shape  | {o ∈ P, s ∈ ⌂, θ ∈ G}         | Shape with origin and rotation |
| Label  | {o ∈ P, t ∈ S, α, β ∈ G}      | Graphic label with rotation    |


## Drawing Elements

Graphic elements are composite data types.

| Name | Signature                       | Meaning              |
|------|---------------------------------|----------------------|
| CRT  | {x, y ∈ Q}                      | Cartesian Point      |
| POL  | {r ∈ P, θ ∈ G}                  | Polar Point          |
| VEC  | {o, p ∈ CRT}                    | Vector               |
| CRC  | {o ∈ CRT, r ∈ P}                | Circle               |
| ARC  | {o ∈ CRT, r ∈ P, θ₁, θ₂ ∈ G}    | Arc                  |
| SQR  | {o ∈ CRT, b ∈ P, θ ∈ G}         | Square with rotation |
| TRG  | {a, b, c ∈ CRT, θ₁, θ₂, θ₃ ∈ G} | Triangle             |
| REG  | {o ∈ CRT, n, r ∈ P, θ ∈ G}      | Regular Shape        |
| PLG  | {v ∈ [VEC]}                     | Polygon Shape        |


## Drawing Keywords

| Keyword | Description                |
|---------|----------------------------|
| draw    | Create a shape on a layer  |
| wipe    | Remove shapes from a layer |
| show    | Show canvas                |
| hide    | Hide canvas                |


**Read next:** [System Library](/projects/bee/library/)
