# TODO
1. Rethink DSL + timing approach with relative integer beats instead of float seconds 
2. Rewrite parser to have better error handling for better user expirience
3. Rethink output approach - output should be in stdout
4. Develop tablike grid sequencer DLS addition (we will keep both)
5. Parser should now duration of last played sequence as well
6. Do we need so many commands like '>' '~' '@', or we can use '~' and '@' to tell its type isnt a note

--- questions
do I really need repeat line operation?


## DSL development
As a reference I am going to use SonicPI API
https://gist.github.com/carltesta/424cc9e42f4de2ed52a41a612e22dc69
Combining with what I learned about using GuitarPro 5

### List of special characters appear on keyboard
§±!@#$%^&*()-_=+
[]{}
;:'"\|
`~,<.>/?

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
Its damn hell trying to calculate pauses and it has visual overhead
just to play two notes sequentially 0.5 sec each.
For ex: 
```
> e3 0.5 : 0.5 > d3 0.5 : 0.5
```

#### Take 1 - shrink
```
>e3 0.5 :0.5 >d3 0.5 :0.5
```
Or shrink even more
```
>e3 0.5:0.5 >d3 0.5:0.5
```
Rule:
separate -
- operations by special chars
- arguments by whitespaces
- and newlines only needed if || is used

#### Take 2 - default delay duration
```
>e3 0.5: >d3 0.5:
```
Rule
value for ":" defaults to
- duration of last used operation (> ~ or @)

#### Take 3 - default delay duration + offset
```
>e3 0.5:-0.2 >d3 0.5:-0.2
or even
>e3 .5:-.2 >d3 .5:-.2
```
Rule
same as in Take 2 but we can mod it by using + and - sign

#### Take 4 - extended math with defaults
Make math work not only in sequences but in seq args list
Allow all 4 operators +-/*
Rule
8/2 - div 8 by 2
x/2 - div x by 2 (x going to be substituted during seq arguments substitution)
/2 - div prev op nth arg by 2


## I want to have live-coding
How can we do this?


## Builing new parser dev-notes
#### Current state
helpers:
    - modOpArg - apply op on a and b (floats)
    - splitOpArgs - split string into lhs rhs and op, or return as it if no operators found
    - getOpNoteArg - get freq where lhs is note
    - getOpArg - get freq where lhs is freq
and how freq is parsed:
freq - getOpArg
midi - MidiToFreq(int getOpArg)
note - MidiToFreq(int getOpNoteArg)

so expanded it looks like:
- freq
    getOpArg:
        splitOpArgs
        modOpArg (lhs, rhs, op)
- midi
    MidiToFreq(int of
        getOpArg:
            splitOpArgs
            modOpArg (lhs, rhs, op)
    )
- note 
    MidiToFreq(int of
        getOpNoteArg:
            splitOpArgs
            NoteToMidi
            modOpArg (lhs, rhs, op)
    )


## Brainstorming unordered playback issue
Background:
    I wan Bleeder to support streaming into "ffplay"

Here is my test input:
```
[lane.main] 
>120
@first 1 _1
>180 _2 
@second 80
>220

[lane.first]
args d:1
@second _d
@second 

[lane.second] 
args m:60 
>m _2 >m+10 
```

Using reverse shared buffer.
processing main... 
saving into buffer:
midi    delta   note
120     0
180     1       because _1 after @first
220     2       because _2 after >180

processing first...     (line: @first 1 _1)
nothing to save into buffer (and nothing to emmit yet, because no time advance "_")

processing second...    (line: @second _d)
saving into buffer:
60  0
"_2" found:
    flush everything before advancing time by 2
    flushed:
    120 0
    60  0
    180 1
advance time

...hm, I think we cant flush it...
What if we going to have:
[lane.third]
args d:1 m:60
>m _d >m+10


>120
@third 2 100
@third 1 200
_1
>300

buffer gets
120 0
300 1

going to third 2 100
put 100 0
see _d which is _0 - and that means that we are going to flush 120 and 100
but there is 200 that needs to sound at the same time.
argh!





Output going to be according to Aelin:
main parse loop:
- >120 → emit t=0, midi=120 
- @first 1 → buffer: {t=0,60}, {t=1,60}, {t=2,70}, {t=3,70} 
- _1 → main_t=1, drain t≤1: emit {t=0,60}, {t=1,60}; buffer: {t=2,70}, {t=3,70} 
- >180 → emit t=1, midi=180 
- _2 → main_t=3, drain t≤3: emit {t=2,70}, {t=3,70}; buffer: empty
- @second 80 → insert {t=3,80}, {t=5,90}; buffer: {t=3,80}, {t=5,90}
- >220 → emit t=3, midi=220 
- end → flush: {t=3,80}, {t=5,90}

final table:

┌─────┬──────┬──────┬─────┬──────┐
│  #  │ midi │ absT │ dt  │ note │
├─────┼──────┼──────┼─────┼──────┤
│ 1   │ 120  │ 0    │ 0   │ C9   │
├─────┼──────┼──────┼─────┼──────┤
│ 2   │ 60   │ 0    │ 0   │ C4   │
├─────┼──────┼──────┼─────┼──────┤
│ 3   │ 60   │ 1    │ 1   │ C4   │
├─────┼──────┼──────┼─────┼──────┤
│ 4   │ 180  │ 1    │ 0   │ —    │
├─────┼──────┼──────┼─────┼──────┤
│ 5   │ 70   │ 2    │ 1   │ A#4  │
├─────┼──────┼──────┼─────┼──────┤
│ 6   │ 70   │ 3    │ 1   │ A#4  │
├─────┼──────┼──────┼─────┼──────┤
│ 7   │ 220  │ 3    │ 0   │ —    │
├─────┼──────┼──────┼─────┼──────┤
│ 8   │ 80   │ 3    │ 0   │ G#5  │
├─────┼──────┼──────┼─────┼──────┤
│ 9   │ 90   │ 5    │ 2   │ F#6  │
└─────┴──────┴──────┴─────┴──────┘






## blabla about riff
what if this
```
b>>-
-a-a
```
split into:
```
line[0]=
b
>
>
-
line[1]=
-
a
-
a
```





# Developing a final version (for now)
Operators `> < @ $ _ | :`

How it can be easy to convert into IR?
Lane example:
`>c4:2 |<+7:1 |<+5:1`
Which can be formatted into set of:
```
>c4:2   // T=0 Ta=2
|       // in case of | - do not advance time by Ta
<+7:1   // >c4+7:1
|
<+5:1   // >c4+7+5:1
```

Riff example:
```
c4 >
c4+7 _
c4+12 _
```
Which can be formatted into set of:
>c4:2
|
>c4+7
|
>c+12

Or if we will have two parsers, it should just rotate the matrix into
c4 c4+7 c+12
> _ _

But do we need to rotate? Because how can we know that c4 duration is 2?
Also lets imaging using "<" operator in Riff. Its much easier to parse it sequentially cell by cell, then go to next row.


# Dev chPrev
we have stringified varsions of previous ins properties
we have getArg with fallback
we have evalArg which calculates and returns number

what we want
`>40 <  ` - 40 40
`>40 <+7` - 40 47
`>40 <60` - 40 60

so, we have stringified 40
in case `<  ` - do 40
in case `<+7` - do 40+7
in case `<60` - do 60

we can think of it like
`<  ` - 40 + 0
`<+7` - 40 + 7
`<60` - 60

or maybe
`<  `
`<+7`
`<60`

# Developing Riff syntax
We got tokens
```
{"40", "_", "c4", "_"},
{"80", "88", "_", "68"},
```
We expect
```
"m40.0 v1.0 d1.0 t0.0",
"m80.0 v1.0 d1.0 t0.0",
"m88.0 v1.0 d1.0 t1.0",
"m60.0 v1.0 d1.0 t2.0",
"m68.0 v1.0 d1.0 t3.0",
```
So, to generate it correctly, we need to read by columns













# Developing live-coding

## Ideas for live-coding feel
1.  Repeat updated
    `@seq1 < < <` - user updates seq1 - repeat updated
2.  Live audition
    `send seq2` - user sends seq2 - seq2 is played in parallel with current track
### AI ideas
1.  Multi track 
    `play mute solo stop` - similar to Live audition - dont think we need it
2.  Quantized triggering
    `play lead --on-beat` - we can have Live audition work this was too
3.  Pattern queueing
    `next chorus` - replace next sequence with chorus
4.  Live parameter tweaks
    `set bass volume` - going to be hard to implement, can be done through file edits
### AI TUI idea
┌─ bleeder live ─────────────────────┐
│ ● main    @intro      [====----]   │
│ ● bass    @bassline   [========]   │
│ ○ drums   (muted)     [====----]   │
│                                    │
│ BPM: 120  Time: 0:32  CPU: 12%     │
└────────────────────────────────────┘
Kinda interesting. But we need to make app aware of boundaries

## What I saw during live-loop using Strudel
- Multiple tracks (called orbits I guess)
- Updates happens on repeats
- Tweak params like patch, volume, ADSR etc in the real time
- What else? Dont know yet

So in Bleeder its like:
```
# in main
@track1 |
@track2

# in tracks
@seq1 < < <
```
And thats it, I feel like Bleeder can do this all.
As for tweaking params in real-time - we can do it with Vibe - no time tracking needed here, only fast parser

Here is what I "composed" in Strudel:
```js
const lpf = slider(2755.6,400,3000)
const tr = slider(6,0,12,2)

$: note("b3 c4 a3 e4 d4 a3 e3 <<g4 f3> <d5 e5>>")
    .decay(0.2)
    .delay(0.7)
    .sound("saw")
    .lpf(lpf.add(100))
    .transpose(tr)

$: note("b2 <c3 d3>")
    .decay(0.3)
    .delay(1.2)
    .sound("square")
    .lpf(lpf.add(20))
    .transpose(tr)
```

And how bleed analog going to look like
```toml
[vibe.lead]
transpose = 6
lpf = 2855
delay = 0.7
decay = 0.2
patch = 'saw'

[vibe.bass]
transpose = 6
lpf = 2775
delay = 1.2
decay = 0.3
patch = 'square'

[lane.main]
content = '''
$lead @track1:g4 | $bass @track2:c3 <d3 <c3 <d3
$lead @track1:d5 | $bass @track2:c3 <d3 <c3 <d3
$lead @track1:f3 | $bass @track2:c3 <d3 <c3 <d3
$lead @track1:e5 | $bass @track2:c3 <d3 <c3 <d3
'''

[lane.track1]
vars = 'end:g4'
content = '>b3 >c4 >a3 >e4 >d4 >a3 >end'

[lane.track2]
vars = 'end:c3'
content = '>b2 >end'
```

## How Bleeder live-coding going to work
On every update - reparse whole IR
Before update we played Instruction with idx 20
Q: where to put idx after update?



## How can we know boundaries


## Command to play using ffplay
```sh
bleeder play -seq main '/Users/pburi/DarkGrayCat/golang/bleeder/experiments/test.toml' | ffplay -autoexit -nodisp -loglevel quiet -
```



## devnotes about new WAV rendering
HOW THIS WORKS FOR STREAMING:

One-shot (file with correct header):
wav := NewWAV2(44100, 1)
samples := renderAll(irp)
wav.WriteHeader(len(samples), file)
wav.WriteSamples(samples, file)

Streaming (header once, then chunks):
wav := NewWAV2(44100, 1)
wav.WriteHeader(0, os.Stdout)  // Size 0 or huge for streaming

for {
    chunk := renderChunk(irp, pos, duration)
    wav.WriteSamples(chunk, os.Stdout)
    pos += duration
}

# MAIN devnote about rendering flow
```sh
bleeder song.bleed | bleeder-wav | ffplay -
bleeder song.bleed | bleeder-midi | daw
bleeder song.bleed | bleeder-irp | tui-viz
```

```sh
bleeder song.bleed | tee \
  >(bleeder-wav | ffplay -) \
  >(bleeder-midi | daw) \
  >(bleeder-irp | viz)
```

Is current Bleeder design plan is fully unix-way?
No. It's not.

We're building a monolith:
- Parses DSL
- Generates IR
- Renders to multiple formats (WAV/MIDI/IRP)
- Manages playback state
- Runs a server

Unix way is:
bleeder: .bleed → IR (ONE JOB)

Then separate tools:
bleeder song.bleed | bleeder-wav | ffplay -
bleeder song.bleed | bleeder-irp | tui-viz
bleeder song.bleed | tee >(bleeder-wav | ffplay -) >(bleeder-irp | tui-viz)

Bleeder should JUST output IR as text. That's it.
Renderers are separate programs. Live-coding is watchexec. Parallel outputs are tee.
No renderers inside bleeder. No streaming loops. No goroutines. No BleedCtx complexity.
Just: parse .bleed → write IR text → exit.
