# Issue: Operator Overload Ambiguity (Geometric Congruence)

The triple-bar operator (`≡`) is currently underspecified or potentially colliding with other uses. In geometric contexts, it is traditionally used to denote geometric congruence (e.g., similar shape, same angles, same number of sides, potentially different scale/size).

## Impact
- **Semantic Overload:** The operator needs an explicit home in the geometric specification to avoid conflict with equality or assignment.
- **Formal Definition:** Bee currently lacks a definition for geometric "congruence" as distinct from "equality."

## Requirements
1. **Define Geometric Congruence:** Use `≡` to represent geometric similarity (same shape, angles, side counts; scale invariance allowed).
2. **Standardize:** Formally map `≡` to geometric types and operations in the geometry specification.
3. **Spec Update:** Update `spec/13-graphics.md` (or equivalent geometry spec) to reserve `≡` for this purpose.
