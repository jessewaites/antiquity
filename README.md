# Antiquity

https://github.com/user-attachments/assets/f6078aad-cb45-4230-8e28-3e4118ede52c

Antiquity is a CLI that scaffolds an opinionated workspace for hunting
through historical sources with AI agents: the folder conventions, the rules
of the game in `AGENTS.md`, the templates, and the traps that burned the
investigations that came before. Your own coding agent does the plumbing.

It grew out of a real investigation, in which an AI pipeline over digitised
colonial archives turned up a lost meteorite fall and a few other things the
catalogues had missed. The process, the mistakes and the lessons are written
up in [AI found a lost meteorite in the archives](https://jessewaites.com/blog/post/ai-found-a-lost-meteorite-in-the-archives).
Antiquity is that process, packaged.

```sh
go install github.com/jessewaites/antiquity@latest
antiquity new bird-sightings
```

That opens the title screen, asks four questions (name, the one-line
question, a longer description, optional API keys) and creates the case.
Everything after the name is optional: ctrl+s skips the rest, and the
generated AGENTS.md tells the agent to ask you for whatever is missing.

```
bird-sightings/
  AGENTS.md          rules, the seven-step loop, traps, folder map, your case
  CASE.md            question, status, plan
  HANDOFF.md         start here next session
  NARRATIVE.md       the story for the write-up, with credits
  lessons.md         traps found in this case's sources
  keys.example.yml   model roles and provider keys (copy to keys.yml)
  sources/ data/ queries/ judges/ runs/ candidates/ evidence/
  findings/ theories/ catalogues/ controls/ todo/ drafts/ templates/
```

The case is `git init`-ed with a pre-commit hook that refuses `keys.yml` and
anything that looks like an API key.

Non-interactive, for agents and scripts (only the name is required):

```sh
antiquity new bird-sightings --yes \
  --question "Pre-1800 bird sightings in Dutch colonial records" \
  --description "Searching ship logbooks for dodo sightings after 1660."
```

Flags: `--no-intro`, `--no-git`, `--dir PATH`. `antiquity intro` plays the
title screen on its own; `--snapshot --at 1500` prints one frame.

## The process it encodes

1. Question, plus a known-answer control.
2. Retrieve wide (regex and embeddings).
3. Sort cheap with a judge model.
4. Read carefully with a reader model.
5. Verify on the original page image. Nothing is a find before this.
6. Novelty check against the catalogues.
7. Save evidence, write it up.

Full rules and the list of traps are in the generated `AGENTS.md`.

## Development

```sh
go run . new test-mystery
go test ./...
```

Go 1.27, Bubble Tea v2, Bubbles v2, Lip Gloss v2.
