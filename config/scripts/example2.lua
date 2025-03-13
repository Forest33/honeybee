local hb = require("honeybee")
local json = require("json")

local socket_on = false
local min_temperature = 23
local max_temperature = 24

function Init()
    hb.newTicker("example ticker", 1000000000)
    return {
        Name = "example2",
        Description = "An example of the script #2",
        Subscribe = {}
    }
end

function OnTicker(name, data)
    print("global variable: ", hb.getGlobal("GlobalVar"))
    hb.deleteGlobal("GlobalVar")
    print("global variable: ", hb.getGlobal("GlobalVar"))
    hb.stopTicker(name)
end
