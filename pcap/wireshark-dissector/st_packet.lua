print("Starting starling packet layer dissector!")

local st_packet = Proto("st_packet","Dissector for the Starling Packet Layer Protocol")

local f_has_data = ProtoField.bool("st_packet.has_data", "Has data", 2, { [1] = "True", [2] = "False"}, 0x80)
local f_continuation = ProtoField.bool("st_packet.continuation", "Continuation", 2, { [1] = "True", [2] = "False"}, 0x40)
local f_length = ProtoField.uint16("st_packet.length", "Packet Length", base.RANGE_STRING, {{0, 1048576, ""}}, 0x3FFF)
local f_data = ProtoField.bytes("st_packet.data", "Data", base.SPACE)

st_packet.fields = { f_has_data, f_continuation, f_length, f_data }


function st_packet.dissector(buffer,pinfo,tree)
    length = buffer:len()
    if length == 0 then return end
  
    pinfo.cols.protocol = st_packet.name

    local subtree = tree:add(st_packet, buffer(), "Starling Packet Layer")
    subtree:add(f_has_data, buffer(0,1))
    subtree:add(f_continuation, buffer(0,1))
    subtree:add(f_length, buffer(0,2))
    subtree:add(f_data, buffer(2,-1))

    local st_network = Dissector.get("st_network")
    st_network:call(buffer(2,-1):tvb(), pinfo, tree)

end

local eth_table = DissectorTable.get("ethertype")
eth_table:add(0x0000, st_packet)


-- local s = ""
-- local test = Dissector.list()
-- for k, v in pairs(test) do
--     s = s .. k .. ":" .. v .. "\n" -- concatenate key/value pairs, with a newline in-between
-- end
-- print(s)


-- https://mika-s.github.io/wireshark/lua/dissector/2017/11/04/creating-a-wireshark-dissector-in-lua-1.html