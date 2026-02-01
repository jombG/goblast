# Test Selection Strategies

Goblast supports multiple strategies for selecting which tests to run based on code changes.

## Available Strategies

### 1. symbol-only
**Most precise, potentially misses tests**

Runs only tests that directly use changed symbols (detected via go/types).

Example:
```
Changed: calculator.Add()
Runs: TestAdd (uses Add directly)
Skips: TestCalculator (doesn't use Add, even if in same package)
```

**Use when**: You want minimal test runs and are confident in type-based detection.

### 2. package-fallback (default)
**Balanced precision and safety**

- If changed symbols have direct test usages → run those specific tests
- If no direct usages detected in a package → run all tests in that package (fallback)

Example:
```
Package A: Changed Add(), TestAdd uses it → run TestAdd only
Package B: Changed internal helper, no direct usages → run all tests in Package B
```

**Use when**: You want smart selection with a safety net (recommended).

### 3. conservative
**Safest, runs more tests**

Runs all tests in all packages that have any code changes.

Example:
```
Changed: calculator.Add()
Runs: All tests in calculator package
```

**Use when**: Maximum safety is required, or changes are risky.

## Usage

```bash
# Use default strategy (package-fallback)
./goblast

# Use specific strategy
./goblast --strategy=symbol-only
./goblast --strategy=conservative

# Debug what tests are selected
./goblast --debug-selection --dry-run
```

## Strategy Selection Matrix

| Changed Code | symbol-only | package-fallback | conservative |
|--------------|-------------|------------------|--------------|
| Exported function | Tests using it | Tests using it, or all if none | All in package |
| Internal helper | Tests using it | All in package | All in package |
| Type definition | Tests using it | Tests using it, or all if none | All in package |
| Multiple packages | Tests per symbol | Smart per-package | All tests all packages |

## Implementation Details

All strategies use:
- AST-based symbol extraction
- go/types for precise usage detection
- Package-aware test discovery

The strategy only affects **which tests are selected**, not how changes or usages are detected.
