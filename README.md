# tcpip-go

Go と Linux TAP デバイスで、Ethernet / ARP / IPv4 / ICMP / UDP / TCP を最小実装する学習用 TCP/IP プロトコルスタックです。

このプロジェクトは OS の TCP/IP スタックを `net.Listen` や `net.Dial` から使うのではなく、`/dev/net/tun` で作った TAP デバイスから Ethernet フレームを直接読み書きします。

## 必要環境

- Linux
- Go 1.23 以上
- `iproute2`
- `ping`, `arping`, `nc`
- 任意: Wireshark または tcpdump

macOS など Linux 以外でも純粋関数の単体テストは実行できますが、TAP デバイスを使った実行は Linux 専用です。

## セットアップ

```sh
go mod download
go test ./...
```

## TAP デバイス作成コマンド

```sh
sudo ip tuntap add dev tap0 mode tap user "$USER"
sudo ip addr add 192.168.100.1/24 dev tap0
sudo ip link set tap0 up
```

自作スタック側の設定は固定です。

- MAC: `02:00:00:00:00:02`
- IPv4: `192.168.100.2`
- UDP echo: `9000`
- TCP echo: `8080`

## 実行方法

```sh
go run ./cmd/stack
```

環境によっては `/dev/net/tun` の権限により `sudo` が必要です。

```sh
sudo go run ./cmd/stack
```

## ping 確認

別ターミナルで実行します。

```sh
ping 192.168.100.2
```

最初に Linux ホストが ARP Request を送り、自作スタックが ARP Reply を返します。その後 ICMP Echo Request に ICMP Echo Reply を返します。

## UDP 確認

```sh
nc -u 192.168.100.2 9000
```

入力した文字列がそのまま返れば成功です。

## TCP 確認

```sh
nc 192.168.100.2 8080
```

3-way handshake 後、入力した文字列がそのまま返れば成功です。学習用の 1 接続最小実装なので、複数同時接続や再送制御は未実装です。

## Wireshark 確認

`tap0` をキャプチャ対象にします。便利な表示フィルタは以下です。

```text
arp
icmp
udp.port == 9000
tcp.port == 8080
```

## トラブルシューティング

- `open /dev/net/tun: permission denied`: `sudo` で実行するか、TAP 作成時の `user "$USER"` を確認します。
- `ping` 前に ARP が解決しない: `ip link show tap0` と `ip addr show tap0` で up 状態と `192.168.100.1/24` を確認します。
- `Destination Host Unreachable`: TAP の IP アドレス、サブネット、スタック側 IP が一致しているか確認します。
- UDP/TCP だけ動かない: `tcpdump -i tap0 -n -vv` か Wireshark で checksum、port、IPv4 Protocol を確認します。
- TCP 接続後に止まる: ACK 番号、SYN/FIN が 1 byte の sequence number を消費する点を確認します。

## sudo が必要な理由

TAP デバイスは仮想 L2 ネットワークデバイスです。作成や `/dev/net/tun` の操作には通常、管理者権限または適切な capability が必要です。この教材では理解しやすさを優先し、Linux ホスト側の設定は `sudo ip tuntap ...` で行います。

## この教材で実装していないこと

- IP fragmentation
- IPv4 options の実用処理
- TCP 再送、輻輳制御、window scaling
- 複数 TCP 接続
- RST の完全処理
- DHCP, DNS, IPv6
- セキュリティ対策
- 本番利用に耐える RFC 完全準拠

## 参考 RFC

- RFC 826 ARP
- RFC 791 IPv4
- RFC 792 ICMP
- RFC 768 UDP
- RFC 793 TCP

## 教材本文

章ごとの解説は [docs/TUTORIAL.md](/Users/tshimobayashi/sources/go-tcp-tp/docs/TUTORIAL.md) にあります。

