
print("Starting starling network layer dissector!")


local st_network = Proto("st_network","Dissector for the Starling Network Layer Protocol")

local f_packet_type = ProtoField.uint8("st_network.packet_type", "Packet Type", base.DEC, { [1] = "RREQ", [2] = "RREP", [3] = "SESS", [4] = "RERR"})
local f_requestid = ProtoField.uint64("st_network.requestid", "Request ID", base.DEC)
local f_ttl = ProtoField.uint16("st_network.ttl", "TTL", base.DEC)
local f_bitmap = ProtoField.bytes("st_network.bitmap", "Contact Bitmap", base.SPACE)

local f_sessionid = ProtoField.uint64("st_network.sessionid", "Session ID", base.DEC)
local f_nonce = ProtoField.bytes("st_network.nonce", "Nonce", base.SPACE)
local f_auth_tag = ProtoField.bytes("st_network.auth_tag", "Authentication Tag", base.SPACE)

local f_data_size = ProtoField.uint32("st_network.data_size", "Data Size", base.DEC)
local f_data = ProtoField.bytes("st_network.data", "Data", base.SPACE)

st_network.fields = { f_packet_type, f_requestid, f_ttl, f_bitmap, f_sessionid, f_nonce, f_auth_tag, f_data_size, f_data}


function st_network.dissector(buffer,pinfo,tree)

    length = buffer:len()
    if length == 0 then return end
  
    pinfo.cols.protocol = st_network.name

    local subtree = tree:add(st_network, buffer(), "Starling Network Layer")

    local packet_type = buffer(0,1):uint()
    if packet_type == 1 then
        subtree:add(f_packet_type, buffer(0,1))
        subtree:add(f_requestid, buffer(1,8))
        subtree:add(f_ttl, buffer(9,2))
        subtree:add(f_bitmap, buffer(11,256))
    elseif packet_type == 2 then
        subtree:add(f_packet_type, buffer(0,1))
        subtree:add(f_requestid, buffer(1,8))
        subtree:add(f_sessionid, buffer(9,8))
        subtree:add(f_nonce, buffer(17,12))
        subtree:add(f_auth_tag, buffer(29,16))
    elseif packet_type == 3 then
        local data_size = buffer(21,4):uint()

        subtree:add(f_packet_type, buffer(0,1))
        subtree:add(f_sessionid, buffer(1,8))
        subtree:add(f_nonce, buffer(9,12))
        subtree:add(f_data_size, buffer(21,4))
        subtree:add(f_data, buffer(25,data_size))
        subtree:add(f_auth_tag, buffer(25+data_size,16))

        local st_transport = Dissector.get("st_transport")
        st_transport:call(buffer(25,data_size):tvb(), pinfo, tree)
    elseif packet_type == 4 then
        subtree:add(f_packet_type, buffer(0,1))
        subtree:add(f_sessionid, buffer(1,8))
    end
end


-- https://mika-s.github.io/wireshark/lua/dissector/2017/11/04/creating-a-wireshark-dissector-in-lua-1.html