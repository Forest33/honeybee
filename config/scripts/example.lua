local hb = require("honeybee")
local json = require("json")

local Monday = 1
local Tuesday = 2
local Wednesday = 4
local Thursday = 8
local Friday = 16
local Saturday = 32
local Sunday = 64

local socket_on = false
local min_temperature = 23
local max_temperature = 24

function Init()
    hb.newAlarm("example alarm", Monday + Tuesday + Wednesday + Thursday + Friday + Saturday + Sunday, 12, 30, 00, {})
    hb.newTimer("example timer", 1000000000 * 3)
    hb.newTicker("example ticker", 1000000000 * 1)

    return {
        Name = "example",
        Description = "An example of the script",
        Subscribe = {
            "zigbee2mqtt/temperature_1",
            "zigbee2mqtt/socket_1"
        }
    }
end

function OnMessage(topic, data)
    if topic == "zigbee2mqtt/socket_1" then
        socket_on = data.state == "ON"
    elseif topic == "zigbee2mqtt/temperature_1" then
        if data.temperature <= min_temperature then
            if socket_on == false then
                hb.publish("zigbee2mqtt/socket_1/set", json.encode({ state = "ON" }))
            end
        elseif data.temperature >= max_temperature then
            if socket_on then
                hb.publish("zigbee2mqtt/socket_1/set", json.encode({ state = "OFF" }))
            end
        end
    end
end

function OnTimer(name, data)
    print("timer called: ", name)
end

function OnTicker(name, data)
    print("ticker called: ", name)
    hb.stopTicker(name)
end

function OnAlarm(name, data)
    print("alarm called: ", name)
end
