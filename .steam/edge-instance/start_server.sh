#!/bin/bash
# Replace with your actual server parameters
./srcds_run \
    -game csgo \ 
    -console \
    -usercon \
    -tickrate 128 \
    +sv_setsteamaccount "replace_with_your_GSLT_token" \ 
    +game_type 0 \
    +game_mode $CSGO_GAME_MODE \
    +mapgroup $CSGO_MAPGROUP \
    +map $CSGO_MAP \
    +hostname "$CSGO_SERVER_NAME" \
    +rcon_password "$CSGO_RCON_PASSWORD"
