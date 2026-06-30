local ESX = exports['es_extended']:getSharedObject()
local WebhookSecret = GetConvar('raven_webmarket_secret', 'change-fivem-secret')
local ApiBase = GetConvar('raven_webmarket_api', 'http://127.0.0.1:8080')

SetHttpHandler(function(req, res)
    if req.path ~= '/raven-webmarket/deliver' or req.method ~= 'POST' then
        res.writeHead(404)
        res.send('')
        return
    end
    if req.headers['X-Webhook-Secret'] ~= WebhookSecret then
        res.writeHead(401)
        res.send('{"error":"unauthorized"}')
        return
    end
    local body = req.body or ''
    local payload = json.decode(body)
    if not payload or not payload.identifier then
        res.writeHead(400)
        res.send('{"error":"invalid payload"}')
        return
    end
    local xPlayers = ESX.GetExtendedPlayers()
    local delivered = false
    for _, xPlayer in pairs(xPlayers) do
        if xPlayer.identifier == payload.identifier then
            for _, item in ipairs(payload.items or {}) do
                xPlayer.addInventoryItem(item.name, item.count)
            end
            TriggerClientEvent('raven-webmarket:notify', xPlayer.source, 'You received items from the web shop!')
            delivered = true
            break
        end
    end
    if not delivered then
        MySQL.insert('INSERT INTO web_mailbox (identifier, discord_id, payload, source_type, source_ref, status) VALUES (?, ?, ?, ?, ?, ?)', {
            payload.identifier,
            payload.discord_id or '',
            json.encode(payload.items or {}),
            payload.source_type or 'order',
            payload.source_ref or '',
            'pending'
        })
    end
    res.writeHead(200)
    res.send('{"status":"ok"}')
end)

CreateThread(function()
    while true do
        Wait(60000)
        local players = ESX.GetExtendedPlayers()
        for _, xPlayer in pairs(players) do
            MySQL.query('SELECT id, payload FROM web_mailbox WHERE identifier = ? AND status = ?', {
                xPlayer.identifier, 'pending'
            }, function(rows)
                if rows then
                    for _, row in ipairs(rows) do
                        local items = json.decode(row.payload)
                        if items then
                            for _, item in ipairs(items) do
                                xPlayer.addInventoryItem(item.name, item.count)
                            end
                            MySQL.update('UPDATE web_mailbox SET status = ?, claimed_at = NOW() WHERE id = ?', {
                                'claimed', row.id
                            })
                            TriggerClientEvent('raven-webmarket:notify', xPlayer.source, 'You have mail rewards from the web shop!')
                        end
                    end
                end
            end)
        end
    end
end)
