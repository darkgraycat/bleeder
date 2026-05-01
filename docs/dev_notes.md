## DSL development
As a reference I am going to use SonicPI API
https://gist.github.com/carltesta/424cc9e42f4de2ed52a41a612e22dc69
Combining with what I learned about using GuitarPro 5

### User scenarios
1. Play file
`bleeder play song.toml`
2. Play stdin
`cat song.toml | bleeder play`
3. Play part using Neovim cmd
`:%w !bleeder play`


### Bleed.toml format examples
#### Example 1: positional arguments
```toml
[seq.riff_1]
content = """
play c3 1 0.5 # command note duration volume
sleep 1
play d3 1 0.6 # play chord d3+e3
play e3 1 0.5 # play chord d3+e3
sleep 1
"""
repeat = 4

[seq.main]
content = """
play @riff_1
sleep 8 # riff_1 duration is 2 * 4 repeats
play @riff_1
"""
repeat = 2
```

#### Example 2: using operators
```toml
[seq.riff_1]
content = """
> c3 1 0.5
~ 1
> d3 1 0.6 > e3 1 0.6
~ 1
"""
repeat = 4

[seq.main]
content = """
> @riff_1
~ @riff_1
> @riff_1
"""
repeat = 2
```

#### Example 3: using operators extended
```toml
[seq.riff_1]
args = "note vol"
repeat = 4
content = """
> :note 1 :vol
. 1 # ~ is going to be used for waves
> :note+2 1 :vol+0.1 | :note+4
. 1
"""

[seq.main]
content = """
> @riff_1 c3 0.5
. @riff_1
> @riff_1 c4 0.5
"""
repeat = 2
```

#### Example 4: real operators with math (too confusing)
```toml
[seq.riff_1]
args = 'note vol'
repeat = 4
content = """
note * vol : 1
: 1
(note + 2 * vol + 0.1) + (note + 4 * vol + 0.1)
: 1
"""

[seq.main]
content = """
riff_1(c3 0.5)
: @riff_1
riff_1(c4 0.5)
"""
repeat = 2
```

#### Example 5: custom operators with minimal syntax
```toml
[seq.riff_1]
args = note vol
repeat = 4
content = """
> note 1 vol : 1
> note+2 1 vol+0.1 | +2 : 1
"""

[seq.main]
repeat = 2
content = """
> @riff_1 c3 0.5 : @riff_1
> @riff_1 c4 0.5 : @riff_1
"""

```

### Flow
CLI/Daemon
    ↓
Parser (TOML + DSL → IR)
    ↓
IR (Intermediate Representation)
    ↓
Generator (IR → WAV samples)
    ↓
Player (samples → audio output)

### Flow reexplained
```
bleeder <cmd>
    <play>
        load cfg
        load bleed
        create Player
        create Bleeder
            for each sequence
                create ir.Program
                parse content into instructions
                store in Bleeder
            on <method>
                <intoIRFull>
                    read main sequence
                    create ir.Program
                    parse content into instructions
                    for each sequence reference
                        call intoIRSeq method
                        result merge with initial ir.Program
                    return initial ir.Program
                <intoIRSeq>
                    return stored ir.Program value from Bleeder
                <intoIRRaw>
                    create ir.Program
                    parse content into instructions
                    for each sequence reference
                        use stored ir.Program value from Bleeder
                        merge with initial ir.Program
                    return initial ir.Program
    <serve>
        // TODO
    <send>
        // TODO
```

**General flow**
```
DSL ("> c3 0.5 1")
    ↓
Command (function)
    ↓
Instruction (generic data: tag, freq, duration, etc.)
    ↓
Player (interprets instruction for output format)
    ↓
WAV / MIDI / TABS
```

**Circular reference validation**
Input data:
```toml
[seq.riff1]
> @riff2

[seq.riff2]
> @riff1

[seq.main]
> @riff1
```
Parsing flow:
refs = []
parse main
    store "main" in refs
    confirm "riff1" not in refs
    parse "riff1"
        store "riff1" in refs
        confirm "riff2" not in refs
        parse "riff2"
            store "riff2" in refs
            confirm "riff1" not in refs
                ERROR: Circular dependency detected



### Implementation details
This one
```
> note 1 vol : 1 ||
> note+2 1 vol+0.1 | +2 : 1
```
Can be read as
```
play note 1 vol wait 1
play note+2 1 vol+0.1 repeat +2 wait 1
```

In case it splitted by special chars:
```
>
note 1 vol
:
1
>
note+2 1 vol+0.1
|
+2
:
1
```

But the most "parsable" I think is:
```
> note 1 vol
: 1
> note+2 1 vol+0.1
| +2
: 1
```

Nope, the most parsable is to expand | and || beforehand
```
> note 1 vol : 1 > note 1 vol : 1
> note+2 1 vol+0.1 > note+4 1 vol+0.1 : 1
```
And after that we going to split into
```
> note 1 vol
: 1
> note 1 vol
: 1
> note+2 1 vol+0.1
> note+4 1 vol+0.1
: 1
```

Experiment
```
> c3 || > d3 ||
```
should be expanded as
```
> c3 > c3 > d3
> c3 > c3 > d3
```

Okay, in case we use array of pointers anyway
so we can copy only pointers not values
In this case lets back to split by lines then by whitespace chars:
```
>                       # line 0
note 1 vol
:
1

>                       # line 1
note+2 1 vol+0.1
|
+2
:
1
```
Algo:
remember index of the first instruction in line

Main problem in parsing splitted by emptyspace - next token has unknown sense

But what if:
collect args till next instruction
on next operation - fill prev instruction with args


## Writting Renderers using IRs


## Riffs DSL example (post-MVP)
```toml
[riff.test]
args = 'a:c3 b:e3'
tempo = 128
content = """
--b- --b- --b- --b-
aa-a aa-a aa-a aa-a
"""
```

#### more on it
What if with notes we also can play sequences?
And make Bleeder as nested visualisation tool?
Because with current syntax it looks like SonicPI
Example of it can be found at `./experiments/example_of_new_syntax_idea.toml`

In case we have sequence with nested sequences with different duration each
What duration of "-" going to be?
Is it shortest or longest sequence duration?
```toml
[seq.example]
args = 'a:@one b:@two'
content = """
a-a
b-b
"""
```


## Rethinking current DSL
