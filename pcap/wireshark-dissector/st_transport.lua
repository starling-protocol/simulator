
print("Starting starling transport layer dissector!")


local st_transport = Proto("st_transport","Dissector for the Starling Transport Layer Protocol")

local f_packet_type = ProtoField.uint8("st_transport.packet_type", "Packet Type", base.DEC, { [1] = "DATA", [2] = "ACK"})
local f_sequenceid = ProtoField.uint32("st_transport.sequenceid", "Sequence ID", base.DEC)
local f_data = ProtoField.bytes("st_transport.data", "Data", base.SPACE)

local f_count = ProtoField.uint32("st_transport.count", "Sequence ID Count", base.DEC)


st_transport.fields = { f_packet_type, f_sequenceid, f_data, f_count}


function st_transport.dissector(buffer,pinfo,tree)

    length = buffer:len()
    if length == 0 then return end
  
    pinfo.cols.protocol = st_transport.name

    local subtree = tree:add(st_transport, buffer(), "Starling Transport Layer")

    local packet_type = buffer(0,1):uint()
    if packet_type == 1 then
        subtree:add(f_packet_type, buffer(0,1))
        subtree:add(f_sequenceid, buffer(1,4))
        subtree:add(f_data, buffer(5,-1))
    elseif packet_type == 2 then
        local count = buffer(5,4):uint()

        subtree:add(f_packet_type, buffer(0,1))
        subtree:add(f_sequenceid, buffer(1,4))
        subtree:add(f_count, buffer(5,4))

        for i = 0, (count * 4) - 1, 4 do
            subtree:add(f_sequenceid, buffer(9 + i + 1,4))
        end
    end
end


-- https://mika-s.github.io/wireshark/lua/dissector/2017/11/04/creating-a-wireshark-dissector-in-lua-1.html