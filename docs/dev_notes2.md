# Rewrite plan
1. finish live-coding, dont polish it yet
2. implement new DSL
3. do refactor of Bleeder struct, renderers and overall

# Commands to test
play mode
```sh
```

live mode
```sh
bleeder live -cfg ~/DarkGrayCat/golang/bleeder/config.toml bleeds/test.toml | bleeder-wav | ffplay -f s16le -ar 44100 -
```

# Reinventing DSL
## Behaviour of [] each operator
RULE:
left is outer, right is inner:
[a b] + [x y] -> a+x a+y b+x b+y
```toml
[a b c]
# will be
a b c

[a b] / 2
# will be
a/2 b/2

8 / [a b]
# will be
8/a 8/b

[a b] & c
# will be
a & c b

c & [a b]
# will be (same as previos)
c & a b

[a b] |
# will be
a b a b

[a b] |+2
# will be
a b a+2 b+2

@chord(e2) | [d3 a4]
# will be
@chord(e2) |d3 |a4

[@chord(e2) @chord(a2)] + 2
# will be
@chord(e2+2) @chord(a2+2) 

@chord(e2 3) + [(2 7) (5 4)]
# will be
@chord(e2 3) @chord(e2+2 3+7) @chord(e2+5 3+4)

[a b] + [x y]
# will be
a+x a+y b+x b+y

[a b] * [1 1 7 1]
# will be
a*1 a*1 a*7 a*1
b*1 b*1 b*7 b*1

[a & b] * 2
# will be
a*2 & b*2

[a & b] + [x y]
# will be
a+x a+y & b+x b+y

[a b] + [x & y]
# will be
a+x & a+y b+x & b+y

[a & b] + [x & y]
# will be
a+x & a+y & b+x & b+y

2 + [a & b] * [x y]
# will be
2+a*x 2+a*y & 2+b*x 2+b*y

[x y] * [a & b] + 2
# will be
x*a+2 & x*b+2 y*a+2 y*b+2
# or
2+a*x & 2+b*x 2+a*y 2+b*y
```

RULE:
right is outer, left is inner:
[a b] + [x y] -> a+x b+x a+y b+y
```toml
[a b] / 2
# will be
a/2 b/2

8 / [a b]
# will be
8/a 8/b

[a b] & c
# will be
a&c b&c

c & [a b]
# will be - or we want `c & a b` in both cases?
c&a c&b 

[a b] |
# will be
a b a b
# but it sounds like a special case

[a b] & [x y]
# will be
a&x b&y ?

[a & b] * 2
# will be
a*2 & b*2

[a & b] + [x y]
# will be
a+x & b+x a+y & b+y

[a b] * [1 1 7 1]
# will be
a*1 b*1 a*1 b*1 a*7 b*7 a*1 b*1

3 * [a & b] + [1 2 3]
# will be
3*a+1 & 3*b+1
3*a+2 & 3*b+2
3*a+3 & 3*b+3
```




### Thought on how to implement [] each operator
When [ - than means we start nested loop for each element
When ] - we stop nested loop
What nested loop will do?

So the main question - should we look ahead or look back?
If we can find out how easily we can produce
`a+x a+y b+x b+y` from `[a b] + [x y]`
all other combinations should work.

A RULE:
`[]` - emits values and tokens from inside:
    if value - apply operator if present
    if token - expand as it

Simple case - scalar to group:
```
2 + [a b] 3
```
1. remember left operand `2`
2. remember operator `+`
3. iterate each in group:
    if its value do `left op item` - `2 + a`
    if its token - insert as it - `2+[a & b]` -> `2+a & 2+b`
4. just insert `3` at the end

Complex case
```
2 + [a & b] * [x y]
```
Solution 1: add into group
```
[2+a 2+b] * [x y]
[2+a*x 2+a*y 2+b*x 2+b*y]
```

**RULE**
LEFT is the entire prefix and
RIGHT is the entire suffix.
Then if either side is a collection, distribute.

`a + b * [x & y] - c`
L - `a+b*`
R - `-c`
OUT:
```
a+b* x -c
&
a+b* y -c
```

`2 + [a & b] * [x y]`
L - `2+[a & b]*`
R - ` `
OUT:
```
2+[a & b]*x
2+[a & b]*y
```
L - `2+`
R - `*x`, `*y`
OUT
```
2+a*x
&
2+b*x
2+a*y
&
2+b*y
```

What we should split:
an expression left of group and right of group

`2 + [a & b] * [x y] - c d`
solution with no-recurse
`2+` - mem
`2+a & 2+b` - mem
`2+a*x & 2+b*x 2+a*y & 2+b*y` - mem
`2+a*x-c & 2+b*x-c 2+a*y-c & 2+b*y-c` - out
its like memorizing what you have in the left side (in array of tokens)
and by moving forward:
    - if its `- c` - append for each item in mem
    - if its `* [x y]` - append for each item in mem for each value in group
in case we hit no operators after `c` - that mean we can "flush" what we have in memory and continue with `d`
DETAILED flow
```
cur     mem
2       2
+       2+
[a      2+a
&       2+a &
b]      2+a & 2+b
*       2+a* & 2+b*
[x      2+a*x & 2+b*x
y]      2+a*x & 2+b*x 2+a*y 2+b*y
-       2+a*x- & 2+b*x- 2+a*y- 2+b*y-
c       2+a*x-c & 2+b*x-c 2+a*y-c 2+b*y-c
```

#### ANOTHER APPROACH
What if I can read it and build a pattern of substitutions?
Like:
`2 + [a b]` expansion pattern is `2+$1 2+$2`
I mean - to have a single pattern which later going to be substituted?
Like 2 slices:
1 - structure pattern
2 - substitution values

`2 + [a b] - c` is
`2+$1-c 2+$2-c` and values `a b`

`[a & b] + [x y]` is
`$1+$3 & $2+$3 $1+$4 & $2+$4` and values `a b x y`

WRONG, correct ones
`2 + [a b] - c` is
`2 + $1 - c` values `a b`

`2 + [a & b] + [x y] * 3` is
`2 + $1 + $2 * 3` values `a & b` and `x y`

So idea is:
iterate and collect expression into "mem"
on `[]` save group tokens into separate [][]string(or map) add a marker
when expression is done
use collected groups to generate substitute values
for
`2 * [a & b] + [0 1 2]`
template is:
`2 * $1 + $2`
groups:
$1 - `a & b`
$2 - `0 1 2`
expected result:
```
2*a+0 & 2*b+0
2*a+1 & 2*b+1
2*a+2 & 2*b+2
```
so, how can we build it?


So for:
```
3 * [a & b] + [1 2 3]
# will be
3*a+1 & 3*b+1
3*a+2 & 3*b+2
3*a+3 & 3*b+3
```
Logs are:
```
TOK     3 * [ a & b ] + [ 1 2 3 ]
GR      [[a & b] [1 2 3]]
TEMP    3 * $1 + $2
```
So we need to iterate as:
```
a 1
&
b 1

a 2
&
b 2

a 3
&
b 3
```

3 groups:
```
TOK     [ a & b ] * [ x y ] - [ 1 2 ]
GR      [[a & b] [x y] [1 2]]
TEMP    $1 * $2 - $3
```
will be
```
a*x-1 & b*x-1
a*y-1 & b*y-1

a*x-2 & b*x-2
a*y-2 & b*y-2
```

```
1 x a
1 x & - emit just &
1 x b
1 y a
```


#### YET ANOTHER LOOK AT PROBLEM
The huge problem is that Im trying to expand without knowing if its an expression
I feel like I need to split not by rows, but by rows+expressions.
But, we agreed to treat `&` as a last char in a row to make next row in parallel.
Why do I need rows during IR generation?
Maybe or tokenizer should return [][][]string then?
```
[
    [ [a+1], [b+3], &],
    [ [a+1], [b+3], &],
]
```
The good part here is that every expression means one instruction.
In other words - we need to split content by instructions anyway.

Okay `&` at the end is a special case. Do we really need it?
I mean do we really need a grid-like?
Because having combined power of `[]` and `@`... Yeah. Lets forget about it.
So, `&` - is basically - skip time adjustment of previos instruction (AT - prev.T).
Se we really need to tokenize by expressions.
What is expression then?
`a + b` is expression
`a & b` I think isnt
So expression is basically can be converted into instruction.



Or what if we can do another level of recurse?
`d 3*[a&b]+[1 2 3]`
Algo: parse char by char, but when hit `[` - recurse
Recursion will look like:




### Main reqirements
- User can put anything into `[]`
- Expansion process should be plain and simple, and work as subtitution
- Group `[a b]` means will be expanded into two sequential
- Group `[a & b]` will be expanded into two parallel





# Proposed file structure
- bleed.go - filetype
- bleeder.go - orchestration and API
- parser.go - IR generation
