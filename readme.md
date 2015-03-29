#dfa

This is a simple renderer for deterministic finite automaton. Currently it's not super beautiful (couple hour hack), but it works.

![](http://i.imgur.com/pjH8y2A.png)

## Usage

 * Clone this repository and run `go get && go build`, or just grab a binary from the bin folder in this repo.
 * Running dfa from the command line with `dfa -input="dfa.txt" -output="vis.svg"` will read the "dfa" file and output the specified svg.

## Input Format

The input format is from my CSC237 class, and is defined thusly:

>Assign a name to each state, with the start state named s0 [this does not matter here!], and the other states named s1, s2,. . . etc. The first line of the text file should have the form `accept si` where si is the name of one of your states. If you have more than one accept state, write them all on that first line, separated by spaces, e.g. accept s0 s2. The remaining lines each specify a single transition, and have the form `si d sj` where [d is a symbol] and si and sj are names of states.

Example:

```
accept s4
s0 0 s1
s0 1 s0
s1 0 s2
s1 1 s0
s2 0 s2
s2 1 s3
s3 0 s1
s3 1 s4
s4 0 s4
s4 1 s4
```

## License

Copyright 2015 by Connor Peet, licensed under the [MIT License](http://opensource.org/licenses/MIT).
