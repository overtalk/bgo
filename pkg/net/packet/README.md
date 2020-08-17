# Packet

- 旨在提供网络包的抽象，具体的解包规则可自定义

## packet buffer
- `packetbuffer.go`
- 用于从 `io.Reader` 中读出一个数据包

## 已有包结构
- `basic_packet.go`
- `tunnel_packet.go`