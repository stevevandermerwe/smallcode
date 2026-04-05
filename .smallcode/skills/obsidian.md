# Obsidian Skill

This is used to interact with obsidian.

## Purpose
Use obsidian to write or read documentation.

## Usage
Obsidian has a cli an to run it simply run `obsidian` in a bash command.   Use the CLI to do most things on Obsidian (ask for help on the cli by running `obsidian help`). Here are some documented example of use.

```
// Open today's daily note
obsidian daily

// Add a task to your daily note
obsidian daily:append content="- [ ] Buy groceries"

// Search your vault
obsidian search query="meeting notes"

// Read the active file
obsidian read

// List all tasks from your daily note
obsidian tasks daily

// Create a new note from a template
obsidian create name="Trip to Paris" template=Travel

// List all tags in your vault with counts
obsidian tags counts

// Compare two versions of a file
obsidian diff file=README from=1 to=3

## Note Conventions

Every note should include frontmatter:

```yaml
---
title: Note Title
date: YYYY-MM-DD
tags:
  - category/subcategory
---
```

Use nested tags (`#ai/sdlc`, `#software-engineering/testing`) to keep the tag tree organized. Link between notes using `[[Note Name]]` wikilinks rather than duplicating content.
