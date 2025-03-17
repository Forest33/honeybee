local hb = require("honeybee")

function Init()
    return {
        Name = "example2",
        Description = "An example of the script #2",
        Subscribe = {},
        Disabled = true
    }
end

function Main()
    hb.newTicker("example ticker", 1000000000)
end

function OnTicker(name, data)
    print("global variable: ", hb.getGlobal("GlobalVar"))
    hb.deleteGlobal("GlobalVar")
    print("global variable: ", hb.getGlobal("GlobalVar"))
    hb.stopTicker(name)
end
