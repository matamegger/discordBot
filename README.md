# discordBot

It's yet not sure in which direction this project will go. Appreciate any constructive input.

## Dependencies

### Go
- [discordgo](https://github.com/bwmarrin/discordgo)
- [go-logging](https://github.com/op/go-logging)
- [dca](https://github.com/bwmarrin/dca)

### System
- [ffmpeg](https://ffmpeg.org/) / opus

[dca](https://github.com/bwmarrin/dca) (and its dependencies ffmpeg/opus) are used to dynamically encode and add new sounds to the "soundboard" function.

## Installing

This assumes you already have a working Go environment, if not please see
[here](https://golang.org/doc/install) first.

`go get` *will always pull the latest released version from the master branch.*

```sh
# Installing the bot
go get github.com/matamegger/discordBot
go install github.com/matamegger/discordBot
```

[discordgo](https://github.com/bwmarrin/discordgo) and [go-logging](https://github.com/op/go-logging) should be installed automatically.
However, you have to install [dca](https://github.com/bwmarrin/dca) and its dependencies manually.

## Starting/Configurating

Assuming you have added your go/bin folder to your path and used go install for discordBot.

```sh
discordBot -t DiscordBot.Token -o discordIdOfTheOwner
```
To get the DiscordBot Token go to [this page](https://discordapp.com/developers/applications/me).
<br>To get your ID type \\@yourname in a server channel, but use only the number (remove the <@>).<sup>*1.</sup>

## Footnotes
1. You could also start the bot, without an owner ID, and ask it with `!get id` for your ID.