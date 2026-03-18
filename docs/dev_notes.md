## DSL development
As a reference I am going to use SonicPI API
https://gist.github.com/carltesta/424cc9e42f4de2ed52a41a612e22dc69
Combining with what I learned about using GuitarPro 5

### An examples
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


### API v0.1
`>` - play command. arguments: freq, dur, vol
`:` - wait command. arguments: time or sequence reference
`|` - repeat last command. arguments: depends on last command (override)
`||` - repeat whole line. arguments: same as for `|`
`@` - reference to sequence or sample
`+ - * /` - modifiers. can be applied to freq, dur, vol

### API v0.1 detailed
`>`  Play command
     Args: freq, dur, vol
     Example: > c3 1 0.5

`:`  Wait command
     Args: time OR sequence reference
     Example: : 1.0
     Example: : @intro

`|`  Repeat last command
     Args: override previous args
     Example: > c3 1 0.5
              | +2      (= > c3+2 1 0.5)

`||` Repeat whole line
     Args: same as |
     Example: > c3 0.5 | +4 : 1 || +7

`@`  Reference to sequence or sample
     Example: > @kick
     Example: : @intro


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
