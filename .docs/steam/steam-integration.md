# SteamAPI User Data Policy & Token Usage

## Overview
ReplayAPI uses the SteamAPI to authenticate users using a secure OAuth 2.0 flow provided by Steam. This document outlines the data that ReplayAPI collects from the SteamAPI and how it is used.

## Verification
When receiving the access token using by underlying clients, ReplayAPI verifies the token with the SteamAPI to ensure its validity and authenticity before proceeding with any further actions on behalf of the user.

## Token Usage
The access token provided by Steam is used to authenticate the user and retrieve necessary information from the SteamAPI. Tokens are NEVER stored in the server. The token is not shared with any third parties and is only used for the purpose of authenticating the user and retrieving steam identification described below.

# Security
# DEPRECATED (TOKEN IS ALREADY BEING DISCARDED AFTER VALIDATION AND NEVER PERSISTED)
The access token provided by Steam contains only `read` scope permissions, which means that ReplayAPI can only **READ** the user's profile information (if public) and **CANNOT** make any changes to their account or perform any other actions on their behalf.

# (DEPRECATED) no need for personal in .club or in community/opensource level
## Basic Information 
### Name
- **Field Name**: `name`
- **Type**: String
- **Description**: The user's display name on Steam.
- **Example**: `"dashdev"`

### Email (if public)
- **Field Name**: `email`
- **Type**: String
- **Description**: The email associated with the user's Steam account.
- **Example**: `"76561199680993440@steamcommunity.com"`

### Image
- **Field Name**: `image`
- **Type**: String (URL)
- **Description**: The URL of the user's profile image.
- **Example**: ![User Image](https://avatars.steamstatic.com/616bba1f7b8dd6adca4fe7ddd1b1af81933642a2_full.jpg)

## Steam Profile Information

### Steam ID (REQUIRED for ReplayAPI)
- **Field Name**: `steamid`
- **Type**: String
- **Description**: The unique Steam ID associated with the user.
- **Example**: `"76561199680993440"`

### Community Visibility State
- **Field Name**: `communityvisibilitystate`
- **Type**: Integer
- **Description**: Indicates the visibility state of the user's Steam community profile.
- **Values**:
    - `1`: Private
    - `2`: Friends Only
    - `3`: Public
- **Example**: `3` (Public)

### Profile State (if provided)
- **Field Name**: `profilestate`
- **Type**: Integer
- **Description**: Indicates whether the user has a Steam profile.
- **Values**:
    - `0`: No profile
    - `1`: Has a profile
- **Example**: `1` (Has a profile)

### Persona Name (if public)
- **Field Name**: `personaname`
- **Type**: String
- **Description**: The user's persona name (displayed on their profile).
- **Example**: `"dashdev"`

### Profile URL (if provided)
- **Field Name**: `profileurl`
- **Type**: String (URL)
- **Description**: The URL to the user's Steam community profile.
- **Example**: [Steam Profile URL](https://steamcommunity.com/id/d_ash_io/)

### Avatar Images
- **Field Name**: `avatar`, `avatarmedium`, `avatarfull`
- **Type**: String (URL)
- **Description**: URLs for different sizes of the user's avatar image.
- **Examples**:
    - Avatar: ![Avatar](https://avatars.steamstatic.com/616bba1f7b8dd6adca4fe7ddd1b1af81933642a2.jpg)
    - Avatar Medium: ![Avatar Medium](https://avatars.steamstatic.com/616bba1f7b8dd6adca4fe7ddd1b1af81933642a2_medium.jpg)
    - Avatar Full: ![Avatar Full](https://avatars.steamstatic.com/616bba1f7b8dd6adca4fe7ddd1b1af81933642a2_full.jpg)

### Real Name (if public)
- **Field Name**: `realname`
- **Type**: String
- **Description**: The user's real name (if provided).
- **Example**: `"dash"`

### Primary Clan ID (if provided)
- **Field Name**: `primaryclanid`
- **Type**: String
- **Description**: The unique ID of the user's primary Steam group (clan).
- **Example**: `"103582791429521408"`

### Time Created
- **Field Name**: `timecreated`
- **Type**: Unix timestamp (seconds since epoch)
- **Description**: The timestamp when the user's Steam account was created.
- **Example**: `1714325308`

### Persona State Flags (if provided)
- **Field Name**: `personastateflags`
- **Type**: Integer
- **Description**: Flags indicating additional persona state information.
- **Example**: `0`

## Expiration Date  (token expiry date) 
- **Field Name**: `expires`
- **Type**: ISO 8601 timestamp
- **Description**: The date and time when this data expires.
- **Example**: `"2024-05-31T00:10:22.567Z"`
