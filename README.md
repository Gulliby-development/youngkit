<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# youngkit

[![CI](https://github.com/Gulliby-development/youngkit/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/youngkit/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`youngkit` is an offline, exact-integer command-line toolkit for integer partitions, Young diagrams and Young tableaux. It is intended for students and researchers in combinatorics, and for anyone who wants to calculate an exact count or apply a standard bijection without writing a program. It uses Go's standard library and makes no network calls.

## Install

Build the command from the repository root with Go:

```sh
go build -o youngkit .
```

This creates the `youngkit` executable in the current directory.

## Usage

The command has twelve subcommands. Their usage forms are:

```text
youngkit count N [--max-part K] [--distinct]
youngkit list N [--max-part K] [--limit L]
youngkit rank SHAPE
youngkit unrank N RANK
youngkit conjugate SHAPE
youngkit hooks SHAPE [--entries K]
youngkit compare A B
youngkit neighbors SHAPE
youngkit rsk ENTRY... | youngkit rsk -
youngkit unrsk P Q
youngkit syt SHAPE [--index K] [--limit L]
youngkit core K SHAPE
```

For `rsk`, provide permutation entries as arguments, or use `-` as the sole argument to read whitespace-separated entries from standard input. Options are written after the command's positional arguments and may also be interspersed with them.

## Shapes and tableaux

A shape is written as comma-separated, positive, nonincreasing parts, for example `5,3,3,1`. The notation `0` denotes the empty shape. Parts must be canonical integers: malformed notation such as a leading-zero part (`03`), a trailing comma, or an empty shape string is rejected.

A run of equal parts may be written `PARTxCOUNT` with the lowercase letter `x`: `4,3x2,1` is `4,3,3,1`, `2x3` is `2,2,2` and `3x1` is `3`. PART and COUNT are canonical integers, COUNT runs from 1 to 5000, and the total size limit is unchanged. Every command that takes a shape accepts the run notation and prints the parts spelled out; `0` stays the only spelling of the empty shape, so `0x1` is rejected as a zero part.

A tableau is written with slash-separated rows and comma-separated entries, for example `1,3,5/2,4/6`; `0` denotes the empty tableau. Standard tableaux contain each integer from 1 through the tableau's size exactly once, with entries increasing from left to right in every row and from top to bottom in every column. Empty rows, malformed entries and tableaux that fail these conditions are rejected.

Counts and ranks are exact integers. Shape and tableau commands accept `0` for their empty object; its printed form is also `0`.

## Commands

`count` counts partitions of a nonnegative integer, optionally limiting the largest part and requiring distinct parts.

`list` prints the first partitions in reverse lexicographic order, with an optional largest-part bound and result limit.

`rank` reports a shape's zero-based position among the partitions of its size and the total number of such partitions.

`unrank` returns the shape at a zero-based partition position, using an exact integer rank.

`conjugate` reflects a shape across its diagonal and prints its Ferrers diagram.

`hooks` prints every cell's hook length and the exact standard-tableau count; `--entries` also requests the semistandard-tableau count for that entry bound.

`compare` compares two equal-size shapes by dominance and reports their partial sums.

`neighbors` lists the shapes formed by adding or removing one box.

`rsk` applies Robinson–Schensted to a permutation and reports its insertion and recording tableaux, along with its longest increasing and decreasing subsequence lengths.

`unrsk` validates two standard tableaux of the same shape and recovers their permutation by inverse Robinson–Schensted.

`syt` counts and lists standard Young tableaux in the documented order; `--index` jumps to a position and `--limit` bounds the number shown.

`core` computes a shape's `k`-core, weight and runner quotient shapes using the abacus.

Examples:

`youngkit hooks 4,3,1 --entries 3` exits with 0 and prints:

```text
shape       4,3,1
size        8
hook        6 4 3 1
hook        4 2 1
hook        1
syt         70
entries     3
ssyt        15
```

`youngkit rsk 2 5 1 4 3` exits with 0 and prints:

```text
perm        2 5 1 4 3
shape       2,2,1
lis         2
lds         3
p           1 3
p           2 4
p           5
q           1 2
q           3 4
q           5
```

`youngkit core 3 5,4,1` exits with 0 and prints:

```text
shape       5,4,1
k           3
core        2,1,1
weight      2
quotient    0=0
quotient    1=1
quotient    2=1
```

## Exit codes

A successful command exits with status `0`. A command or input error exits with status `2`, writes one `youngkit: `-prefixed diagnostic followed by LF to standard error, and writes nothing to standard output. Successful output lines end with LF.

## Layout

- `main.go` - The executable entry point passes process arguments and standard streams to the CLI.
- `internal/base` - Shared input, option, integer, text and error primitives live here.
- `internal/shape` - Shape parsing, diagrams, dominance and Young's lattice operations live here.
- `internal/count` - Partition generation, ranking and exact hook-based counts live here.
- `internal/tableau` - Tableau operations, Robinson–Schensted and abacus calculations live here.
- `internal/cli` - The twelve commands and their dispatch live here.

## Development

Run these checks from the repository root:

```sh
gofmt -l .
go vet ./...
go test ./...
```

The project uses only the Go standard library, so these checks require no external service or runtime dependency.

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED
