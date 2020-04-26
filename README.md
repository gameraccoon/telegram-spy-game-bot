# telegram-spy-game-bot
Helper bot for playing offline games where some information should be destributed between all players except one.

## How to use it
1. Ask all the players join the bot
1. Before the round start ask the players that are participating to press Ready
1. Either type the theme to send it, or press Send Theme to send a random theme from the list

After that the theme will be sent to all the players except one. This player will receive "You are the spy" message.

## Install

In order it to work you need to create `config.json` with this content
```json
{
	"defaultLanguage" : "en-us",
	"extendedLog" : false,
	"availableLanguages" : [
		{"key": "en-us", "name": "English"}
	]
}
```
and `telegramApiToken.txt` that containts telegram API key for your bot.


Run this script to build
```
#!/bin/bash
bot_dir=github.com/gameraccoon
bot_name=telegram-spy-game-bot
bot_exec=${bot_name}
go fmt ${bot_dir}/${bot_name}
go vet ${bot_dir}/${bot_name}
go test -v ${bot_dir}/${bot_name}/...
go install ${bot_dir}/${bot_name}
cp ${GOPATH}/bin/${bot_name} ./${bot_exec}
rm -rf "./data"
cp -r ${GOPATH}/src/${bot_dir}/${bot_name}/data ./
```
and this script to run
```
bot_name=telegram-spy-game-bot
bot_exec=${bot_name}
mkdir -p logs
./${bot_exec} 2>> logs/log.txt 1>> logs/errors.txt & disown
```
