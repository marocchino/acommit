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
