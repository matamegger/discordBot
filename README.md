# discordBot

## Dependencies

- [discordgo](https://github.com/bwmarrin/discordgo)
- [dca](https://github.com/bwmarrin/dca)
- ffmpeg

## Installing

This assumes you already have a working Go environment, if not please see
[here](https://golang.org/doc/install) first.

`go get` *will always pull the latest released version from the master branch.*

```sh
go get github.com/matamegger/discordBot
go install github.com/matamegger/discordBot
```

## Usage

Assuming you have added your go/bin folder to your path and used go install for discordBot.

```sh
discordBot -t DiscordBot.Token -o discordIdOfTheOwner
```
To get the DiscordBot Token go to [this page](https://discordapp.com/developers/applications/me).
To get your id type \@yourname in a server channel, but use only the number (remove the <@>)
