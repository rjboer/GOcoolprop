# Analysis of Brent Solver Rewrite

## Summary of Changes

You've completely rewritten the Brent solver using a different variant of the algorithm. Here's what changed:

### ✅ **Improvements**

1. **Better Initial Setup**
   - Added handling for exact roots at endpoints (`fa == 0` or `fb == 0`)
   - Ensures `|fa| >= |fb|` at initialization so `b` starts as the better approximation
   - Added maximum iteration limit (1000) to prevent infinite loops

2. **Clearer Algorithm Structure**
   - Uses `mflag` to track whether the previous step was bisection or interpolation
   - More explicit acceptance conditions (`cond1` through `cond5`)
   - Better comments explaining each step

3. **Improved Convergence Check**
   - Uses both function tolerance (`|fb| <= tol`) and interval tolerance (`|b-a| <= 2*delta`)
   - Dynamic tolerance calculation

4. **Correct Bookkeeping**
   - Properly maintains history: `d <- c`, `c <- b`
   - Correctly updates bracket `[a,b]` to maintain opposite signs
   - Re-swaps `a` and `b` after each iteration to keep `b` as best approximation

### ⚠️ **Potential Issues**

1. **Infinite Loop Problem**
   - The tests are hanging, suggesting the algorithm may not be converging in some cases
   - Possible causes:
     - The acceptance conditions might be too strict
     - The `a > b` swap (lines 53-56) might be interfering with the algorithm
     - The convergence check might not be triggering

2. **Swap Logic Concern**
   ```go
   // For the acceptance tests we assume a < b; enforce that
   if a > b {
       a, b = b, a
       fa, fb = fb, fa
   }
   ```
   This swap is done BEFORE the acceptance tests, but it might break the invariant that `|fa| >= |fb|`. The algorithm assumes `b` is the better approximation, but this swap doesn't preserve that property.

3. **Acceptance Condition Issues**
   - `cond1`: Checks if `s` is outside `[(3a+b)/4, b]`, but this assumes `a < b`
   - After the swap at line 53-56, if `a` and `b` get swapped, this condition might not work correctly

## Comparison with Original Implementation

| Aspect | Original | Your Rewrite |
|--------|----------|--------------|
| Algorithm Variant | Classic Brent (Numerical Recipes style) | Wikipedia/Modern variant with mflag |
| Swap Strategy | Swap to ensure `\|fc\| < \|fb\|` | Swap to ensure `\|fa\| >= \|fb\|` |
| Acceptance Tests | Simple comparison with `min1`, `min2` | Five explicit conditions |
| Max Iterations | None (infinite loop risk) | 1000 (good!) |
| Convergence | `\|xm\| <= tol1` OR `\|fb\| < tol` | `\|fb\| <= tol` OR `\|b-a\| <= 2*delta` |

## Diagnosis of Hanging Issue

The hanging is likely caused by the `a > b` swap interfering with the algorithm's invariants. Here's why:

1. After updating the bracket (lines 73-80), you ensure `|fa| >= |fb|` (lines 83-86)
2. This might result in `a > b` in some cases
3. Then the next iteration swaps them back (lines 53-56)
4. This creates a cycle where the algorithm never converges

## Recommended Fixes

### Option 1: Remove the `a > b` Swap
The acceptance conditions should work regardless of whether `a < b` or `a > b`. Just adjust `cond1`:

```go
// Don't swap based on a > b
// Instead, adjust cond1 to work for both cases
minAB := math.Min(a, b)
maxAB := math.Max(a, b)
cond1 := (s < (3.0*minAB+maxAB)/4.0) || (s > maxAB)
```

### Option 2: Maintain `a < b` Throughout
If you want to maintain `a < b`, don't swap at the end to ensure `|fa| >= |fb|`. Instead, just track which endpoint has the smaller function value.

### Option 3: Use the Classic Algorithm
The original Numerical Recipes style algorithm (which I was trying to implement) doesn't require maintaining `a < b`. It just ensures the best approximation is in a specific variable.

## Testing Recommendation

Add debug output to see where it's getting stuck:

```go
for i := 0; i < maxIter; i++ {
    fmt.Printf("[iter %d] a=%e, b=%e, fa=%e, fb=%e\n", i, a, b, fa, fb)
    // ... rest of loop
}
```

This will show if it's oscillating between states or just not converging.

## Conclusion

Your rewrite uses a more modern, well-documented variant of Brent's method with better structure and comments. However, the `a > b` swap is likely causing the hanging issue. I recommend either:

1. Removing that swap and adjusting the acceptance conditions, OR
2. Not re-swapping to maintain `|fa| >= |fb|` at the end of each iteration

The algorithm should work once this conflict between the two swap operations is resolved.
