# Mattermost DM Control Plugin

This plugin provides granular control over Direct Messages (DMs) and Group Messages (GMs) in Mattermost, allowing administrators to restrict these features based on user roles and other criteria.

## Features

- Disable DMs and GMs for new users upon joining
- Disable DMs and GMs for existing users
- Allow specific roles to send and receive DMs
- Allow specific roles to participate in group chats
- Exclude specific users from DM and GM restrictions
- Provides an API endpoint for checking user restrictions

## Installation

1. Go to the [releases page](https://github.com/itseasy21/mattermost-dm-control-plugin/releases) of this repository and download the latest release.
2. Upload this file in the Mattermost System Console under **System Console > Plugins > Plugin Management**.
3. Enable the plugin in the System Console.

## Configuration

Navigate to **System Console > Plugins > DM Control** to configure the plugin. The following settings are available:

- **Disable DMs for New Users**: If enabled, direct messages will be disabled for new users when they join.
- **Disable Group Messages for New Users**: If enabled, group messages will be disabled for new users when they join.
- **Disable DMs for Existing Users**: If enabled, direct messages will be disabled for all existing users unless they have an allowed role.
- **Disable Group Messages for Existing Users**: If enabled, group messages will be disabled for all existing users unless they have an allowed role.
- **Excluded Users**: A comma-separated list of usernames that are excluded from DM and GM restrictions.
- **Allowed Roles**: A comma-separated list of roles allowed to send DMs and participate in group chats.

## Usage

Once configured, the plugin will automatically enforce the set restrictions. Users will be notified if they attempt to send a message that is restricted.

## API

The plugin provides an API endpoint to check a user's current restrictions:

```
GET /plugins/com.mattermost.dm-control-plugin/restrictions
```

This endpoint returns a JSON object with the following structure:

```json
{
  "canSendDMs": boolean,
  "canReceiveDMs": boolean,
  "canParticipateInGroupChats": boolean
}
```

## Development

To build the plugin:

```
make
```

This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server:

```
dist/com.mattermost.dm-control-plugin.tar.gz
```

## Reporting Issues

If you encounter any issues or have feature requests, please file them in the [issue tracker](https://github.com/itseasy21/mattermost-dm-control-plugin/issues).

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
