## Bleed file format
Example:
```toml
[meta]
main = 'main'
tempo = 128

[seq.main]
content = '''
@chord5 a3 8 _
@chord5 f3 8 _
@chord5 d3 8 _
@chord5 d3 8 7 _
'''

[seq.chord5]
args = 'note:e2 d:8 w:0 v:1.0'
content = '''
:note d v_w |+7 |+5
'''
```

### Sequence content syntax
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

