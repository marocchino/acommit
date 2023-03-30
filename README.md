# Acommit

Generate commit message with chatgpt api

## Install

Download the binary from [GitHub Releases](https://github.com/marocchino/acommit/releases/) and drop it in your $PATH.

Or use go install

```bash
go install github.com/marocchino/acommit
```

## Usage

```bash
# from your repo
git add .
acommit
```

## Config

You can customize the prompt by modifying `~/.config/acommit/prompt.txt`.

### I18n example

If you want to write the commit message in a different language, you can add that language after the prompt. Let's say you're in Japanese.

```
And, Translate it to Japanese except prefix.
```

### commit convention

Using `gitmoji convention with emoji` by default. You may have several other options.

- [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)
- gitmoji convention with emoji markup
