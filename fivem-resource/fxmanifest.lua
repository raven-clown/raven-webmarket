fx_version 'cerulean'
game 'gta5'

name 'raven-webmarket-bridge'
description 'FiveM bridge for Raven Webmarket delivery and mailbox'
version '1.0.0'

server_scripts {
    '@oxmysql/lib/MySQL.lua',
    'server/main.lua'
}

client_scripts {
    'client/main.lua'
}

dependencies {
    'oxmysql',
    'es_extended'
}
