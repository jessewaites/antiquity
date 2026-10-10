# Antiquity

https://github.com/user-attachments/assets/f6078aad-cb45-4230-8e28-3e4118ede52c

Antiquity is a CLI that scaffolds an opinionated workspace for hunting
through historical sources with AI agents: the folder conventions, the rules
of the game in `AGENTS.md`, the templates, and the traps that burned the
investigations that came before. Your own coding agent does the plumbing.

## Why this exists

In October 2026 the historian Benjamin Breen used AI to search the
transcribed archive of the Dutch East India Company and found a new
eyewitness account of the dodo. I read his post and thought: I can do this,
and I can do it at scale. A week later a single GPU in my office and a chain
of AI models had read millions of pages from that archive, two centuries of
Dutch and American newspapers, and a stack of ship logbooks. They surfaced a
meteorite fall in India in 1812 that no catalogue records, three Javan
rhinoceroses shipped as royal gifts in 1738 to 1740 that never reached the
king, and three volcanic eruptions the Smithsonian's list doesn't have. The
full story is in
[I Pointed AI at 400 Years of Historical Archives](https://jessewaites.com/blog/post/i-pointed-ai-at-400-years-of-archives).

I built the plane as I flew it. The folder layout, the chain of cheap and
expensive models, the rule that nothing is a find until it has been read on
the original page: all of it was invented mid-flight, and most of the
mistakes were made at least once before they became rules. With the
investigation done, Antiquity packages that process so the next person starts
where I finished. It is for anyone with a question about the past, a public
archive that might hold the answer, and a coding agent such as Claude Code or
Codex to do the work.

## What it does

Antiquity does not search archives or call AI models itself. It creates a
case folder with everything an AI coding agent needs to run the investigation
properly, and a briefing that stops it from making the mistakes I made. You
open your agent inside the folder and it takes it from there, asking you for
whatever is missing.

```sh
go install github.com/jessewaites/antiquity@latest
antiquity new my-investigation
cd my-investigation
# initialize your agent
```

That opens the title screen, asks four questions (a name, the one-line
question, a longer description, optional API keys) and creates the case.
Everything after the name is optional: ctrl+s skips the rest, and the
generated AGENTS.md tells the agent to ask you for whatever is missing.

```
my-investigation/
  AGENTS.md          the briefing: rules, the loop, traps, folder map, your case
  CASE.md            the question, status, and a plan the agent starts on
  HANDOFF.md         start here next session
  NARRATIVE.md       the story for the write-up, with credits
  lessons.md         traps found in this case's sources
  keys.example.yml   model roles and provider keys (copy to keys.yml)
  sources/ data/ queries/ judges/ runs/ candidates/ evidence/
  findings/ theories/ catalogues/ controls/ todo/ drafts/ templates/
```

The case is `git init`-ed with a pre-commit hook that refuses `keys.yml` and
anything that looks like an API key.

## The process it encodes

This is the loop from the investigation, written into every generated
`AGENTS.md`. Each step exists because skipping it produced a false find.

1. **Ask a question, and bury a wristwatch.** Before searching for anything
   new, pick a known event the pipeline must re-find. If a metal detector
   can't find the watch you buried in your own yard, you fix the detector
   before you trust its silence anywhere else. Mine had to find Breen's dodo
   and the Laki and Tambora eruptions before it was allowed to look for
   anything unknown.
2. **Retrieve wide.** Nobody in the 1600s spelled anything the same way
   twice, and text recognition adds its own errors, so keyword search misses
   most of what is there. Turn every passage into a mathematical fingerprint
   of its meaning and search by what a passage is about. This is the
   expensive step: millions of passages, one overnight run on a GPU.
3. **Sort cheap.** A search returns tens of thousands of hits. A tiny, fast
   decision model reads every one and answers only narrow questions: is this
   a real animal? is it wild? where is it? Reading 59,000 mentions of
   elephants this way cost about three dollars.
4. **Read carefully.** Only the survivors go to a bigger model, which
   translates them and pulls out dates, places and the exact quote. It sees
   dozens, not thousands.
5. **Verify on the original page.** Open the scan of the handwritten letter
   or the printed newspaper and check the transcription against it. Nothing
   is a find before this step. Until then it is a candidate, and the agent
   is told to call it one.
6. **Check it isn't already known.** Before announcing anything, check the
   catalogues the specialists themselves use, and write down how each one
   was checked.
7. **Save the evidence and write it up.** The full page image and a crop,
   with the archive reference anyone can look up, plus a finding note and a
   line in the narrative recording who had which idea.

The briefing also carries the traps that cost me the most time. The worst is
the dateline: a report printed in Batavia under "here" may describe
something that happened in Ceylon, and the archive's finding aid may file a
Cape Town inventory under Java. Then the word traps: in the Cape records a
"sea cow" is a hippo, and the *Meermin* is a ship, not a mermaid. Storms get
reported as earthquakes. Models misread names and fill the gaps with
confidence. Each of these is a rule now, and the agent is told to add new
ones to `lessons.md` the moment it hits them.

## For agents and scripts

Only the name is required:

```sh
antiquity new my-investigation --yes \
  --question "Pre-1800 bird sightings in Dutch colonial records" \
  --description "Searching ship logbooks for dodo sightings after 1660."
```

Flags: `--no-intro`, `--no-git`, `--dir PATH`. `antiquity intro` plays the
title screen on its own; `--snapshot --at 1500` prints one frame.

## Development

```sh
go run . new test-mystery
go test ./...
```

Go 1.27, Bubble Tea v2, Bubbles v2, Lip Gloss v2.
