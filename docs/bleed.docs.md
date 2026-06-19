## Bleed file format
Example:
```toml
[meta]
main = "main"
tempo = 128
include = ["./vibes.bleed.toml"]

[vibe.synth1]
// TODO
content = """
"""

[lane.main]
content = """
@chord5 a3 8 _
@chord5 f3 8 _
@chord5 d3 8 _
@chord5 d3 8 7 _
"""

[lane.chord5]
args = "note:e2 d:8 w:0 v:1.0"
content = """
:note d v_w |+7 |+5
"""

[riff.song1]
// TODO
content = """
"""
```

### Meta section
// TODO: description
`main`      main sequence name
`tempo`     beath per minute
`include`   included bleed file paths

### Vibe sections
// TODO: description

### Lane sections
// TODO: description
`args`      sequence arguments in format `key:val key2:val2`
`content`   sequence content in LaneDSL format

### Riff sections
// TODO: description
`args`      sequence arguments in format `a:val b:val2`
`content`   sequence content in RiffDSL format

## Content syntax

### Lane section format
`>` Play midi
    Args: midi, duration, volume
    Example: `>60 2 .5`

`:` Play note
    Args: note, duration, volume
    Example: `:c#3 2 .5`

`~` Play frequency
    Args: freq, duration, volume
    Example: `~440.17 2 .5`

`_` Wait before next operation
    Args: 
    Examples: `>60_1`

`|` Repeat last operation
    Args: modify/override args of previous operation
    Examples: `>60 2 |*2 1`

`@` Play sequence by name
    Args: name, arguments defined by sequence
    Example `@chord c5 .5`

`$` Switch vibe (instrument)
    Args: name, arguments defined by vibe
    Example `$bass 1 :e3`

### Riff section format
// TODO

### Vibe section format
// TODO


### Lane content characters
`>` Play note or midi
`@` Play nested sequence
`_` Advance time (silence)
`|` Repeat last play operation
`&` Play in parallel with last one
`$` Switch vibe
