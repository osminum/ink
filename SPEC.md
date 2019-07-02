# Ink language specification

This is the source of truth for the Ink programming language.

## Syntax

Ink has an LR(1) grammar that can be parsed successfully with at most 1 lookahead.

Ink's syntax is inspired by JavaScript and Go, but strives to be minimal. This is not necessarily a comprehensive grammar, but expresses the high level structure and mostly up-to-date with the interpreter implementation.

```
Program: Expression*

Expression: (Atom | BinaryExpr | MatchExpr) ','
ExpressionList: '(' Expression* ')'


Atom: UnaryExpr | EmptyIdentifier
        | Identifier | FunctionCall
        | Literal | ExpressionList

UnaryExpr: UnaryOp Atom

EmptyIdentifier: '_'
Identifier: (A-Za-z@!?)[A-Za-z0-9@!?]*

FunctionCall: Atom ExpressionList

Literal: NumberLiteral | StringLiteral
        | BooleanLiteral | FunctionLiteral
        | ObjectLiteral | ListLiteral

NumberLiteral: (0-9)+ ['.' (0-9)+]
StringLiteral: '\'' (.*) '\''

BooleanLiteral: 'true' | 'false'
FunctionLiteral: (Identifier | '(' (Identifier ',')* ')')
        '=>' ( Expression | ExpressionList )

ObjectLiteral: '{' ObjectEntry* '}'
ObjectEntry: Expression ':' Expression
ListLiteral: '[' Expression* ']'


BinaryExpr: (Atom | BinaryExpr) BinaryOp (Atom | BinaryExpr)


MatchExpr: (Atom | BinaryExpr) '::' '{' MatchClause* '}'
MatchClause: Atom '->' Expression


UnaryOp: (
    '~' // negation
)
BinaryOp: (
    '+' | '-' | '*' | '/' | '%' // arithmetic
    '&' | '|' | '^' // logical and bitwise
    | '>' | '<' // arithmetic comparisons
    | '=' // value comparison operator
    | 'is' // reference comparison operator
    | ':=' // assignment operator
    | '.' // property accessor
)
```

A few quirks of this syntax:

- All variables use lexical binding and scope, and are bound to the most local ExpressionList (execution block)
- Commas (`Separator` tokens) are always required where they are marked in the formal grammar, but the tokenizer inserts commas on newlines if it can be inserted, except after unary and binary operators and after opening delimiters, so few are required after expressions, before closing delimiters, and before the ':' in an Object literal. Here, they are auto-inserted during tokenization.
    - This allows for "minification" of Ink code the same way JavaScript source can be minified. Minified Ink code can be more compact, because in Ink, almost all whitespace is unnecessary (except those wrapping the `is` operator).
- String literals cannot contain comments. Backticks inside string literals are counted as a part of the string literal. String literals are also multiline.
    - This also allows the programmer to comment out a block with an explanation, simply like this:
    ```
    realCode()
    ` this block is commented out for testing reasons
    someOtherCode()
    `
    moreRealCode()
    ```
- List and object property/element access have the same syntax, which is the reference to the list/object followed by the `.` (property access) operator. This means we access array indexes with `arr.1`, `arr.(index + 1)`, etc. and object property with `obj.propName`, `obj.(computed + propName)`, etc.
- Object (dictionary) keys can be arbitrary expressions, including variable names. If the key is a single identifier, the identifier's name will be used as a key in the dict, and if it's not an identifier (a literal, function call, etc.) the value of the expression will be computed and used as the key. This seems like it may cause trouble conceptually, but turns out to be intuitive in practice.
- Assignment is always (re)declaration of a variable in its local scope; this means, for the moment, there is no way to mutate a variable from a parents scope (it'll just shadow the variable in the local scope). I think this is fine, since it forbids a class of potentially confusing state mutations, but I might change my mind in the future and add an assignment-that-isn't-declare. Note that this doesn't affect composite values -- you can mutate objects from a parents scope.
- Ink allows boolean algebra with both logical/bitwise (`&|^`) and algebraic (`+*~`) operators, and which one is used depends on context.
- The only control flow constructs are the function call and the match expression (`a :: {b -> c...}`), and the only control flow construct that branches the execution flow is the match expression. This makes Ink programs simple to analyze programmatically and simple to audit manually.

## Types

Ink is strongly but dynamically typed, and has seven non-extendable types.

- Number
- String
- Boolean
- Null
- Composite (including both Objects (dictionaries) and Lists, like Lua tables)
- Function

Composite and Function types are reference-typed, which means assigning a composite to a variable just assigns a reference to the same composite or function value. All other types are value-typed, which means assigning these values to variables will create new copies of those values. i.e.

```
` for simple values `
a := 3, b := a
a := 42

b = 42 `` false, since assignment of values are all copies


` for composite values `
list := [1, 2, 3]
twin := list
clone := clone(list) `` makes a shallow clone

list.(len(list)) := 4 `` append 4 to list
list.(len(list)) := 5 `` append 5 to list

len(list) = 5 `` true
len(twin) = 5 `` true, since it keeps the same reference
len(clone) = 5 `` false, since it keeps a copy of the value instead
```

These are tested in [samples/test.ink](samples/test.ink).

## Concurrency

Ink achieves concurrenty in two ways, through an event loop and through concurrent Ink programs that communicate via serialized message passing.

Callbacks / event loop and closures is one kind of abstraction over concurrency, and message passing to a completely different execution thread is a different kind of abstraction over concurrency. I think this mirrors two different kinds of concurrency in the real world -- concurrency by way of asynchrony (callbacks, event loop) and concurrency by way of isolation and encapsulation in the problem space (threading). So these are both supported by Ink and used in these different contexts.

### Event loop

A single process of Ink program first executes its entrypoint programs, and then optionally exits to an event loop to respond to system events.

### Concurrent processes

This is behind rationale that a program is fundamentally a representation of a single system evolving sequentially, and shared state means two threads are actually a single program, which breeds all sorts of complexity when a single system tries to mutate in two different sequences. Rust's solution is innovative (compile time static checking that shared mutation never occurs), but a more minimal and Inky way dealing with this is to not have shared state, and only communicate by passing serialized data (messages) between threads of execution that are otherwise spawned and execute in isolation. This is in essence JavaScript workers, but where messages can be any serialized data. 

Ink implements this with three builtin functions, `listen(processID, handler) => null` and `notify(processID, message) => null` for sending and receiving messages, and `spawn(function) => processID` for spawning threads (spawn should be renamed, but not sure what the ideal name is). ProcessID (pid) is an opaque object passed around but it's a valid Ink value/type. Once a function has been spawned off into a separate thread, it can choose to listen. Notify will _not block_ even if nothing is listening (nothing in Ink does unless explicitly documented / chosen). The handler will receive the message as its only argument.

These are the right primitives, but we can build much more sophisticated systems and designs, like a state reducer or a task scheduler, into the standard library as we choose and find useful.

## Builtins

### System interfaces

- `in(callback<string>)`: Read from stdin or until ENTER key (might change later)
- `out(string)`: Print to stdout
- `read(string, number, number, callback<list<number>>)`: Read from given file descriptor from some offset for some bytes
- `write(string, number, list<number>, callback)`: Write to given file descriptor at some offset
- `listen(string, callback<list<number>>)`: Bind to a local TCP or UDP port and start handling requests
- `wait(number, callback)`: Call the callback function after at least the given number of seconds has elapsed
- `rand() => number`: a pseudorandom floating point number in interval `[0, 1)`
- `time() => number`: number of seconds in floating point in UNIX epoch

### Math

- `sin(number) => number`: sine
- `cos(number) => number`: cosine
- `pow(number, number) => number`: power, also stands in for finding roots with exponent < 1
- `ln(number) => number`: natural log
- `floor(number) => number`: floor / truncation

### Type casts and utilities (implemented as native functions)

- `string(any) => string`
- `number(any) => number`
- `len(composite) => number`: length of a list or list-like composite value
- `keys(composite) => list<string>`: list of keys of the given composite

## Standard library

Ink's standard library is under active development, and contains utilities like `map`, `filter`, `reduce`, `clone`, and `slice`. Find the source code in the meantime under [samples/stdlib.ink](samples/stdlib.ink).

## Other implementation notes

- Ink source code is fully UTF-8 / Unicode compatible. Unicode printed non-whitespace characters are valid variable and function identifiers, as well as the characters `?`, `!`, and `@`.
- Ink is fully tail call optimized, and tail calls are the default looping / jump primitive for programming in Ink.