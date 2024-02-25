# Media Bot

## Overview

A Discord Bot for media playback and management. The bot was built
using [discordgo](https://github.com/bwmarrin/discordgo) to interact with the discord API. I created my own client for
interacting with Google's APIs, and I'm using kkdais [youtube](https://github.com/kkdai/youtube/v2) package To download
YouTube videos. Since I plan on deploying this with docker, I created a GitHub workflow to create and publish an image
to dockerhub on merges into the main branch. The bot is capable of sending alerts to a discord channel of your choosing.
To avoid excessive notifications Alerts are only sent when there is a critical failure, or when the bot is taken
offline. All messages sent to the alert channel also show up in logs. You can opt out of receiving alerts by leaving the
ALERT_CHANNEL_ID environment variable blank.

Note: This app is hardly stable and far from feature complete, use at your own risk!

## Usage

- `/download` - Download a YouTube video from a URL or search query
- `/search` - <b>beta</b> | Search for YouTube videos with a given query
- `/watch` - <b>coming soon</b> | Play a YouTube video in a voice channel
- `/listen` - <b>coming soon</b> | Play music in a voice channel

## Configuration

The app requires the following environment variables:

- GUILD_ID: the ID of the server you wish to run the bot in. To obtain the Guild ID you need to enable developer tools.
- APPLICATION_ID: the ID of the bot. For the bot to work, you need to create a bot in the discord developer portal, once
  the bot is created you can copy the application ID and bot token.
- BOT_TOKEN: you can get the bot token by clicking `Reset Token` on the bot page of the discord developer portal.
- ALERT_CHANNEL_ID: specify which channel to send alerts to

These can be set in a `.env` or by using the `export` command.

## Deployment

- Using the [Docker Image](https://hub.docker.com/repository/docker/zachsampson/discord-bot/general):
    - With a .env file
        ```bash
        docker run --name media-bot --env-file=.env zachsampson/discord-media-bot:latest
        ```
    - Manually setting the environment variables:
      ```bash
        docker run --name media-bot zachsampson/discord-media-bot:latest \
          -e GUILD_ID='foo' \  
          -e APPLICATION_ID='bar' \
          -e BOT_TOKEN='baz' \
          -e ALERT_CHANNEL_ID='quz'
        ```

- Without Docker
    - Set the environment variables however your heart desires
      ```bash 
      go run main.go
      ```

## Future Plans

- Play, queue, and control music directly in a discord voice channel
- Play YouTube videos directly in a voice channel
- Use some sort of queue to handle alert messages
- Custom tts in voice channels