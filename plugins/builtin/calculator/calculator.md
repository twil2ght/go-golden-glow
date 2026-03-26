# Calculator Plugin

A built-in plugin for performing arithmetic calculations and numeric comparisons.

## Features

- **Arithmetic Calculations**: Evaluate expressions with `+`, `-`, `*`, `/`, and parentheses
- **Numeric Comparisons**: Compare two values using `<`, `>`, `=`, `<=`, `>=`, `!=`

## Usage

### 1. Arithmetic Calculations

#### Natural Language Triggers

```
calculate 2 + 3 * 4
what is (10 - 5) / 2
compute 100 / 4 + 5
2 + 3 equals
```

#### Node Format

```
[node:extractor] [namespace:calculator] [expression:2+3*4]
```

#### Parameters

| Parameter    | Description                           |
|--------------|---------------------------------------|
| `expression` | The arithmetic expression to evaluate |

#### Response Example

```
User: calculate 2 + 3 * 4
System: The result is 14

User: what is (10 - 5) / 2
System: The result is 2.5
```

### 2. Numeric Comparisons

#### Natural Language Triggers

```
is 5 greater than 3
is 10 less than 20
does 7 equal 7
is 100 larger than 50
is 3 smaller than 5
```

#### Node Format

```
[node:checker] [namespace:calculator] [left:5] [operator:>] [right:3]
[node:checker] [namespace:calculator] [left:10] [operator:<] [right:20]
[node:checker] [namespace:calculator] [left:7] [operator:=] [right:7]
```

#### Parameters

| Parameter  | Values                          | Description         |
|------------|---------------------------------|---------------------|
| `left`     | numeric value                   | Left operand        |
| `right`    | numeric value                   | Right operand       |
| `operator` | `<`, `>`, `=`, `<=`, `>=`, `!=` | Comparison operator |

#### Supported Operators

| Operator | Aliases    | Description           |
|----------|------------|-----------------------|
| `<`      | `lt`       | Less than             |
| `>`      | `gt`       | Greater than          |
| `<=`     | `le`       | Less than or equal    |
| `>=`     | `ge`       | Greater than or equal |
| `=`      | `==`, `eq` | Equal                 |
| `!=`     | `ne`       | Not equal             |

#### Response Example

```
User: is 5 greater than 3
System: Yes, 5 is greater than 3

User: is 10 less than 5
System: No, 10 is not less than 5

User: does 7 equal 7
System: Yes, 7 equals 7
```

## Implementation Details

- **Extractor**: Evaluates arithmetic expressions and returns result as `variable.Item`
- **Checker**: Performs numeric comparisons and returns boolean result
- **Executor**: Not implemented (returns nil)
- **Supported Operations**: Addition, subtraction, multiplication, division, parentheses grouping