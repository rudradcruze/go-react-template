#!/usr/bin/env bash
# Single-row composite statusline:
#   [model effort] - <branch>  <graft status centered>  <tokens flush right>
# Width: Claude Code exports COLUMNS to statusline scripts; /dev/tty is a
# fallback, then a fixed width. Graft's ANSI colors are kept and its extra
# rows dropped. Ambiguous-width glyphs (◤ ⚠ …) count as 2 cells so the line
# cannot render wider than the terminal and clip the right edge.
input=$(cat)

model=$(jq -r '.model.display_name // "?"' <<<"$input")
effort=$(jq -r '.effort.level // ""' <<<"$input")
tokens=$(jq -r '(((.context_window.total_input_tokens // 0) + (.context_window.total_output_tokens // 0)) / 1000 * 10 | round) / 10' <<<"$input")

# Empty outside a git repo or on a detached HEAD, then the " - <branch>" part is omitted.
branch=$(git -C "$(jq -r '.workspace.current_dir // .cwd // "."' <<<"$input")" branch --show-current 2>/dev/null)

left="[$model · $effort]${branch:+ - $branch}"
right="${tokens}k tokens"

# First line of the graft statusline only — its extra rows are dropped, and
# graft's corner glyph goes too (it lives in the installed graft package, not
# ours, so strip it here rather than edit the package).
graft=$(printf '%s' "$input" | node "${CLAUDE_PROJECT_DIR:-.}/.claude/helpers/graft-statusline.cjs" 2>/dev/null | head -n1 | sed 's/◤ //')

# Display-cell width in one node pass: strip ANSI, count codepoints, then add
# a second cell per ambiguous-width glyph (geometric shapes, misc technical/
# dingbats, arrows/brackets — graft's ◤ ⚠ ▸ set). Node is locale-proof.
cells=$(node -e 'const s = require("fs").readFileSync(0, "utf8").replace(/\x1b\[[0-9;]*m/g, "").replace(/\n$/, "");
const wide = (s.match(/[⌀-➿─-◿⬀-⯿]/g) || []).length;
console.log([...s].length + wide);' <<<"$graft" | sed 's/\x1b\[[0-9;]*m//g')
cells=${cells:-0}

cols=${COLUMNS:-$( (stty size < /dev/tty) 2>/dev/null | awk '{print $2}')}
cols=${cols:-100}
# COLUMNS is the raw terminal width, but the statusline renders inside the
# footer's inset — pad to a few cells short so the tail never clips.
cols=$((cols - 4))

avail=$((cols - ${#left} - ${#right} - cells))
gap1=$((avail / 2))
(( gap1 < 1 )) && gap1=1
gap2=$((avail - gap1))
(( gap2 < 1 )) && gap2=1

printf '%s%*s%s%*s%s\n' "$left" "$gap1" '' "$graft" "$gap2" '' "$right"
