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
### Behaviour of [] foreach operator
```toml
[a b]+2
# will be
a+2 b+2

2+[a b]
# will be (same as previos)
2+a 2+b

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
```

### Proposed file structure
- bleed.go - filetype
- bleeder.go - orchestration and API
- parser.go - IR generation
