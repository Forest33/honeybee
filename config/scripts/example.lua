local hb = require("honeybee")
local json = require("json")

local socket_on = false
local min_temperature = 23
local max_temperature = 24

function Init()
    return {
        Name = "example",
        Description = "An example of the script",
        Subscribe = {
            "zigbee2mqtt/temperature_1",
            "zigbee2mqtt/socket_1"
        }
    }
end

function Main()
    hb.newAlarm("example alarm", "", { "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday" }, 9, 40, 00, {})
    hb.newAlarm("example alarm 2", "", { "Mon", "Tues", "Wed", "Thurs", "Fri", "Sat", "Sun" }, 9, 40, 01, {})
    hb.newTimer("example timer", 1000000000 * 3)
    hb.newTicker("example ticker", 1000000000 * 1)

    hb.setGlobal("GlobalVar", "value #1")
    print("check global variable: ", hb.getGlobal("GlobalVar"))
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
    hb.setGlobal("GlobalVar", 33)
    hb.stopTicker(name)
end

function OnAlarm(name, data)
    print("alarm called: ", name)
    --hb.sendMessage("Test message in Telegram")
    --hb.pushNotify("my-super-secret-topic-name", "Message title", "Message text", "high")
end
