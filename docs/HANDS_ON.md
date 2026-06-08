# Go + Linux TAP TCP/IP Stack Hands-on

この文書は、完成コードを読むための説明ではなく、自分で実装するための手順です。

各章では、まず自分で TODO を埋めます。詰まった場合だけ `final` branch の実装を見ます。

```sh
git checkout final
```

戻るとき:

```sh
git checkout main
```

## 第0章: 作るものを理解する

作るもの:

- Linux TAP から Ethernet frame を読む
- ARP Reply を返す
- IPv4 header を parse / marshal する
- ICMP Echo Reply で ping を通す
- UDP echo server を作る
- TCP 3-way handshake と echo server を作る
- TCP FIN を扱う

使わないもの:

- `net.Listen`
- `net.Dial`
- `net.TCPConn`
- `net.UDPConn`
- `net/http`

使うもの:

- `os.File`
- `encoding/binary`
- `golang.org/x/sys/unix`
- `time`
- `fmt`
- `errors`
- `log`

演習:

1. OS の TCP/IP stack と、この教材で作る stack の違いを説明してください。
2. TAP と TUN の違いを調べてください。

## 第1章: TAP デバイスを Go から読む

作る package:

```text
internal/tap
internal/hexdump
cmd/stack
```

実装 TODO:

- `/dev/net/tun` を `unix.Open` で開く
- `TUNSETIFF` ioctl を呼ぶ
- `IFF_TAP | IFF_NO_PI` を指定する
- `os.File.Read` で frame を読む
- 読んだ byte slice を hex dump する

確認:

```sh
ping 192.168.100.2
```

この時点では ping は失敗してよいです。ARP frame が hex dump されれば成功です。

よくあるバグ:

- `IFF_NO_PI` を忘れて先頭 4 bytes が増える
- `tap0` が down のまま
- `/dev/net/tun` の権限がない

演習:

1. 受信 frame の先頭 14 bytes を手で読んでください。
2. EtherType の位置を確認してください。

## 第2章: Ethernet frame を parse する

作る package:

```text
internal/ethernet
```

実装 TODO:

- `MAC [6]byte` を定義する
- `Frame` に `Dst`, `Src`, `EtherType`, `Payload` を持たせる
- `Parse([]byte) (Frame, error)` を作る
- `Marshal(Frame) ([]byte, error)` を作る
- 14 bytes 未満なら error にする
- EtherType `0x0806` と `0x0800` を const にする

確認:

- `ping` 時に EtherType `0x0806` が来る
- ARP 解決後は EtherType `0x0800` が来る

演習:

1. Broadcast MAC を判定する関数を追加してください。
2. short frame の単体テストを書いてください。

## 第3章: ARP Reply を返す

作る package:

```text
internal/arp
```

実装 TODO:

- ARP packet 28 bytes を parse する
- Operation `1` を Request、`2` を Reply として const 化する
- `TargetIP == 192.168.100.2` の Request だけに応答する
- Sender と Target を入れ替えて Reply を作る
- Ethernet の宛先 MAC は Request の送信元 MAC にする

確認:

```sh
arping -I tap0 192.168.100.2
ip neigh show dev tap0
```

演習:

1. 自分宛でない ARP Request を無視してください。
2. Reply の Sender MAC が `02:00:00:00:00:02` であることをテストしてください。

## 第4章: IPv4 header を parse する

作る package:

```text
internal/ipv4
```

実装 TODO:

- Version と IHL を読む
- Total Length を検証する
- Header Checksum を検証する
- SrcIP / DstIP を読む
- Protocol `1`, `6`, `17` を const にする
- `Marshal` で checksum を計算する

確認:

- ARP 解決後、ICMP packet の Protocol が `1` として読める

演習:

1. IHL が 20 bytes 未満の packet を error にしてください。
2. checksum を壊した packet の単体テストを書いてください。

## 第5章: ICMP Echo Reply で ping を通す

作る package:

```text
internal/icmp
```

実装 TODO:

- Type `8` を Echo Request、Type `0` を Echo Reply として扱う
- Identifier / Sequence をそのまま返す
- Payload をそのまま返す
- ICMP checksum を計算する
- Ethernet / IPv4 / ICMP を組み合わせて Reply を返す

確認:

```sh
ping 192.168.100.2
```

演習:

1. Echo Request 以外を無視してください。
2. Wireshark で Identifier と Sequence が一致することを確認してください。

## 第6章: UDP Echo Server を作る

作る package:

```text
internal/udp
```

実装 TODO:

- UDP header 8 bytes を parse する
- SrcPort / DstPort / Length / Checksum を読む
- pseudo header を含む checksum を計算する
- DstPort `9000` の payload をそのまま返す

確認:

```sh
nc -u 192.168.100.2 9000
```

演習:

1. port `9001` 宛を無視してください。
2. UDP checksum の検証を単体テストにしてください。

## 第7章: TCP 3-way handshake を実装する

作る package:

```text
internal/tcp
```

実装 TODO:

- TCP header を parse / marshal する
- SYN / ACK flag を const にする
- TCP checksum を pseudo header 込みで計算する
- `LISTEN -> SYN_RECEIVED -> ESTABLISHED` を実装する
- SYN が sequence number を 1 消費することを反映する

確認:

```sh
nc 192.168.100.2 8080
```

この章では接続成立まででよいです。

演習:

1. SYN 以外を LISTEN 状態で無視してください。
2. SYN-ACK の ACK number を Wireshark で確認してください。

## 第8章: TCP Echo Server を作る

実装 TODO:

- ESTABLISHED 後の payload を読む
- 受信 payload 長だけ ACK number を進める
- 受信 payload をそのまま PSH/ACK で返す
- ACK-only packet には echo を返さない

確認:

```sh
nc 192.168.100.2 8080
```

入力した文字列が返れば成功です。

演習:

1. payload なし ACK に返信しないテストを書いてください。
2. echo を大文字変換に変えてみてください。

## 第9章: TCP connection close を扱う

実装 TODO:

- FIN flag を扱う
- FIN が sequence number を 1 消費することを反映する
- FIN/ACK を返す
- 最後の ACK を受けたら LISTEN に戻す

確認:

```sh
nc 192.168.100.2 8080
```

`nc` 終了後、もう一度接続できれば成功です。

演習:

1. close 後に state が LISTEN に戻るテストを書いてください。
2. FIN の ACK number を Wireshark で確認してください。

## 第10章: Wireshark でデバッグする

便利な filter:

```text
arp
icmp
udp.port == 9000
tcp.port == 8080
```

見るポイント:

- Ethernet Src / Dst MAC
- ARP Sender / Target
- IPv4 Total Length
- IPv4 Header Checksum
- ICMP Checksum
- UDP Checksum
- TCP Seq / Ack
- SYN / FIN が sequence number を 1 消費しているか

演習:

1. checksum をわざと壊して Wireshark の表示を確認してください。
2. TCP ACK number を 1 ずらして Linux 側の反応を観察してください。

## 第11章: 発展課題

- DHCP
- DNS
- TCP retransmission
- TCP window
- 複数 TCP connection
- HTTP server
- IPv6
- fuzzing
- property-based testing

