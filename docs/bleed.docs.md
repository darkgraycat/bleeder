## Bleed file format
Example:
```toml
[seq.riff_1]
args = 'note:e2 vol:1.0'
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

### Sequence content syntax
`>`  Play note
     Args: note, dur, vol
     Example: `> c3 1 0.5`

`~`  Play frequency
     Args: freq, dur, vol
     Example: `> c3 1 0.5`

`@`  Play sequence
     Args: name, ...sequence arguments
     Example: `@ chord c5 0.5`

`:`  Wait command
     Args: time
     Example: `: 1.0`

`|`  Repeat last Play instruction
     Args: override args of previous instruction
     Example: `> c3 1 0.5 | +2`

`||` Repeat whole line
     Args: same as for `|`
     Example: `> c3 0.5 | +4 : 1 || +7`

