# TCP/IP 自作ハンズオン ワークブック

この教材は「読んで終わり」ではなく、章ごとに手を動かして小さな合格条件を満たしていく形式です。

## この教材の使い方

各章でやることは固定です。

1. 章の「観察する packet」を Wireshark か tcpdump で見る
2. 指定されたファイルの TODO を埋める
3. `scripts/check-chapter.sh NN` を実行する
4. Linux TAP 上で確認コマンドを実行する
5. 詰まったら `final` branch の同じファイルを見る

完成コードを丸写しする必要はありません。逆に、全部を白紙から考える必要もありません。`main` branch のスターターコードには、package 構成、型、定数、関数名、エラー、TAP 入出力ループが入っています。

## 事前準備

```sh
go test ./...
```

最初の状態では、通常の `go test ./...` は通ります。章別のテストは build tag 付きなので、次のように明示して実行します。

```sh
scripts/check-chapter.sh 02
```

第2章の TODO を埋める前は失敗します。失敗したテストが、その章で直すべき具体的な仕様です。

Linux TAP の準備:

```sh
sudo ip tuntap add dev tap0 mode tap user "$USER"
sudo ip addr add 192.168.100.1/24 dev tap0
sudo ip link set tap0 up
```

## 第1章 TAP を読む

### 目的

Linux host から届く Ethernet frame を Go の `os.File.Read` で読めることを確認します。

### ここで身に付けること

- TAP は L2 の仮想 NIC である
- `/dev/net/tun` は TUN/TAP を操作する入口である
- `TUNSETIFF` ioctl で fd と `tap0` を紐づける
- `IFF_TAP` を付けると Ethernet frame が読める
- `IFF_NO_PI` を付けると Linux 独自 header が消え、先頭が Dst MAC になる

### 読むファイル

- `internal/tap/tap_linux.go`
- `internal/hexdump/hexdump.go`
- `internal/stack/stack.go`
- `cmd/stack/main.go`

### 実行

```sh
go run ./cmd/stack
```

別 terminal:

```sh
ping 192.168.100.2
```

この時点で ping は成功しません。hex dump が出れば合格です。

### 観察ポイント

ARP Request は Ethernet broadcast なので、先頭 6 byte が `ff ff ff ff ff ff` になります。EtherType は offset 12 から 2 byte で、ARP は `08 06` です。

hex dump は、最初は次のように区切って見ます。

```text
ff ff ff ff ff ff  02 00 00 00 00 01  08 06  ...
宛先 MAC            送信元 MAC          種類    ARP packet
```

この段階では ARP の中身まで読めなくて構いません。まず「Ethernet header 14 byte の後ろに別の packet が続く」という入れ子構造に慣れます。

### よくある詰まり

- `operation not permitted`: TAP 作成または `/dev/net/tun` の権限不足
- hex dump の先頭が `00 00 ...` のように 4 byte ずれる: `IFF_NO_PI` がない
- 何も読めない: `tap0` が down、IP が付いていない、ping 先が違う

## 第2章 Ethernet

### 目的

生の byte slice を Ethernet frame として読めるようにします。

### 実装するファイル

- `internal/ethernet/ethernet.go`

### 実装する関数

- `Parse`
- `Marshal`

### Packet の中身

Ethernet header は 14 byte です。

| offset | Go slice | byte数 | フィールド | 読み方 |
| ---: | --- | ---: | --- | --- |
| 0 | `b[0:6]` | 6 | 宛先 MAC address | そのまま 6 byte を `MAC` に copy |
| 6 | `b[6:12]` | 6 | 送信元 MAC address | そのまま 6 byte を `MAC` に copy |
| 12 | `b[12:14]` | 2 | EtherType | `binary.BigEndian.Uint16` で読む |
| 14 | `b[14:]` | 可変 | Payload | EtherType が示す次の protocol の byte 列 |

実際の ARP Request の先頭は、例えばこう見えます。

```text
ff ff ff ff ff ff  02 00 00 00 00 01  08 06  00 01 ...
```

同じ byte 列を Ethernet header として読むとこうです。

| byte列 | 意味 |
| --- | --- |
| `ff ff ff ff ff ff` | 宛先 MAC。broadcast。LAN 上の全員宛 |
| `02 00 00 00 00 01` | 送信元 MAC。Linux host 側の MAC |
| `08 06` | EtherType。`0x0806` なので ARP |
| `00 01 ...` | Payload。ここから先は ARP packet |

`Payload` は 14 byte 目以降です。EtherType は network byte order なので `binary.BigEndian` で読みます。

EtherType の代表例:

| 値 | 意味 | この教材での扱い |
| --- | --- | --- |
| `0x0806` | ARP | 第3章で parse する |
| `0x0800` | IPv4 | 第4章以降で parse する |

ここで大事なのは、Ethernet は中身を深く理解しないことです。Ethernet は「次は ARP です」「次は IPv4 です」と EtherType で教えて、Payload を次の parser に渡すだけです。

### 実装手順

1. `len(b) < HeaderLen` なら `ErrShortFrame`
2. `copy(f.Dst[:], b[0:6])`
3. `copy(f.Src[:], b[6:12])`
4. `f.EtherType = binary.BigEndian.Uint16(b[12:14])`
5. `f.Payload = b[14:]`
6. `Marshal` では `HeaderLen + len(Payload)` の slice を作って逆順に詰める

`Parse` の完成イメージ:

```go
var f Frame
copy(f.Dst[:], b[0:6])
copy(f.Src[:], b[6:12])
f.EtherType = binary.BigEndian.Uint16(b[12:14])
f.Payload = b[14:]
return f, nil
```

`Marshal` はこの逆向きです。`Frame` の field を byte slice の決まった位置へ戻します。

### 自動チェック

```sh
scripts/check-chapter.sh 02
```

### 実機確認

`stack.HandleFrame` で `ethernet.Parse` を呼び、EtherType を log に出します。`ping` するとまず `0x0806` が見えるはずです。

## 第3章 ARP

### 目的

Linux host の `Who has 192.168.100.2?` に答えます。

### 実装するファイル

- `internal/arp/arp.go`
- `internal/stack/stack.go`

### Packet の中身

Ethernet/IPv4 の ARP packet は 28 byte です。

| offset | Go slice | byte数 | フィールド | この教材での値 |
| ---: | --- | ---: | --- | --- |
| 0 | `b[0:2]` | 2 | HardwareType | `1`。Ethernet |
| 2 | `b[2:4]` | 2 | ProtocolType | `0x0800`。IPv4 |
| 4 | `b[4]` | 1 | HardwareLen | `6`。MAC address は 6 byte |
| 5 | `b[5]` | 1 | ProtocolLen | `4`。IPv4 address は 4 byte |
| 6 | `b[6:8]` | 2 | Operation | `1` Request、`2` Reply |
| 8 | `b[8:14]` | 6 | SenderMAC | 質問している側、または回答する側の MAC |
| 14 | `b[14:18]` | 4 | SenderIP | 質問している側、または回答する側の IP |
| 18 | `b[18:24]` | 6 | TargetMAC | Request では未知なので 0 が多い |
| 24 | `b[24:28]` | 4 | TargetIP | 探したい IP |

Request は Operation `1`、Reply は Operation `2` です。

Linux host が `192.168.100.2` を探すときの意味:

| フィールド | 値の例 | 意味 |
| --- | --- | --- |
| SenderMAC | `02:00:00:00:00:01` | Linux host の MAC |
| SenderIP | `192.168.100.1` | Linux host の IP |
| TargetMAC | `00:00:00:00:00:00` | まだ分からない |
| TargetIP | `192.168.100.2` | 自作 stack の IP |

自作 stack が Reply するときは、Sender と Target の立場を入れ替えます。

### 実装手順

1. `Parse` で 28 byte 未満を error にする
2. `binary.BigEndian` で 2 byte field を読む
3. `Reply` は `Operation == OpRequest` かつ `TargetIP == localIP` のときだけ返す
4. Reply の Sender は自作 stack、Target は request の Sender にする
5. `stack.HandleFrame` で EtherType ARP を `arp.Parse` に渡す
6. `arp.Marshal` した payload を Ethernet frame に包んで返す

### 自動チェック

```sh
scripts/check-chapter.sh 03
```

### 実機確認

```sh
arping -I tap0 192.168.100.2
ip neigh show dev tap0
```

`02:00:00:00:00:02` が見えれば合格です。

## 第4章 IPv4

### 目的

Ethernet payload の中にある IPv4 header を読み、checksum を検証します。

### 実装するファイル

- `internal/ipv4/checksum.go`
- `internal/ipv4/ipv4.go`

### Packet の中身

最小 IPv4 header は 20 byte です。IHL は 32bit word 単位なので、byte 数にするには `* 4` します。

checksum は header だけに対する 16bit one's complement sum です。検証時は、checksum field を含んだ header 全体の `Checksum(header)` が `0` になれば valid です。

この教材では IPv4 options は扱わないので、まず 20 byte header だけを読めれば十分です。

| offset | Go slice | byte数 | フィールド | 読み方 |
| ---: | --- | ---: | --- | --- |
| 0 | `b[0]` | 1 | Version / IHL | 上位4bitが version、下位4bitが IHL |
| 1 | `b[1]` | 1 | TOS | この教材では保存しなくてよい |
| 2 | `b[2:4]` | 2 | Total Length | IPv4 packet 全体の長さ |
| 4 | `b[4:6]` | 2 | ID | 返信時に適当な値を入れる |
| 6 | `b[6:8]` | 2 | Flags / Fragment Offset | fragment は扱わない |
| 8 | `b[8]` | 1 | TTL | 返信では `64` |
| 9 | `b[9]` | 1 | Protocol | `1` ICMP、`6` TCP、`17` UDP |
| 10 | `b[10:12]` | 2 | Header Checksum | IPv4 header の checksum |
| 12 | `b[12:16]` | 4 | SrcIP | 送信元 IP |
| 16 | `b[16:20]` | 4 | DstIP | 宛先 IP |
| 20 | `b[20:totalLen]` | 可変 | Payload | ICMP/UDP/TCP など |

`b[0]` の読み方:

```text
0x45 = 0100 0101
       ---- ----
       4    5
       |    |
       |    IHL。5 * 4 = 20 byte header
       IPv4
```

### 実装手順

1. `Checksum` で 2 byte ずつ足す
2. 奇数 byte が残ったら上位 byte として足す
3. carry を fold する
4. 最後に bit 反転する
5. `Parse` で Version, IHL, Total Length を検証する
6. `Marshal` で checksum field を 0 のまま header を作り、最後に checksum を書く

Protocol field の分岐:

| Protocol | 次に渡す parser |
| ---: | --- |
| `1` | `icmp.Parse` |
| `17` | `udp.Parse` |
| `6` | `tcp.Parse` |

### 自動チェック

```sh
scripts/check-chapter.sh 04
```

### 実機確認

ARP が通った後に `ping` し、Protocol `1`、DstIP `192.168.100.2` の IPv4 packet が読めれば合格です。

## 第5章 ICMP

### 目的

ICMP Echo Request に Echo Reply を返し、ping を成功させます。

### 実装するファイル

- `internal/icmp/icmp.go`
- `internal/stack/stack.go`

### 実装手順

1. ICMP header 8 byte を parse する
2. checksum を検証する
3. Type `8` Code `0` だけ Echo Request として扱う
4. Reply は Type `0`、Identifier/Sequence/Payload は request からコピー
5. Ethernet と IPv4 の送信元/宛先を入れ替えて返す

### Packet の中身

ICMP Echo の header は 8 byte です。

| offset | Go slice | byte数 | フィールド | Echo Request | Echo Reply |
| ---: | --- | ---: | --- | ---: | ---: |
| 0 | `b[0]` | 1 | Type | `8` | `0` |
| 1 | `b[1]` | 1 | Code | `0` | `0` |
| 2 | `b[2:4]` | 2 | Checksum | 計算値 | 計算値 |
| 4 | `b[4:6]` | 2 | Identifier | コピー | Request からコピー |
| 6 | `b[6:8]` | 2 | Sequence | コピー | Request からコピー |
| 8 | `b[8:]` | 可変 | Payload | 任意 | Request からコピー |

ping が期待しているのは、Type だけを `8` から `0` に変え、それ以外の Identifier、Sequence、Payload が対応している Reply です。

### 自動チェック

```sh
scripts/check-chapter.sh 05
```

### 実機確認

```sh
ping 192.168.100.2
```

## 第6章 UDP

### 目的

UDP port `9000` に来た payload をそのまま返します。

### 実装するファイル

- `internal/udp/udp.go`
- `internal/ipv4/ipv4.go` の `PseudoHeader`
- `internal/stack/stack.go`

### 実装手順

1. UDP header 8 byte を parse する
2. Length が 8 byte 以上、受信 byte 数以下であることを確認する
3. checksum が 0 でなければ pseudo header 込みで検証する
4. `DstPort == 9000` のときだけ返信する
5. 返信は SrcPort/DstPort を入れ替え、Payload をコピーする

### Packet の中身

UDP header は 8 byte です。

| offset | Go slice | byte数 | フィールド | 意味 |
| ---: | --- | ---: | --- | --- |
| 0 | `b[0:2]` | 2 | SrcPort | 送信元 port |
| 2 | `b[2:4]` | 2 | DstPort | 宛先 port。echo は `9000` |
| 4 | `b[4:6]` | 2 | Length | UDP header + payload の長さ |
| 6 | `b[6:8]` | 2 | Checksum | pseudo header 込みの checksum |
| 8 | `b[8:length]` | 可変 | Payload | echo する中身 |

UDP checksum では UDP header だけでなく IPv4 pseudo header も足します。pseudo header は「この UDP はどの IP からどの IP へ、どの protocol 番号で、何 byte 送られたか」を checksum に混ぜるための仮想 header です。

### 自動チェック

```sh
scripts/check-chapter.sh 06
```

### 実機確認

```sh
nc -u 192.168.100.2 9000
```

## 第7章 TCP handshake

### 目的

TCP port `8080` で 3-way handshake を成立させます。

### 実装するファイル

- `internal/tcp/tcp.go`
- `internal/tcp/checksum.go`
- `internal/stack/stack.go`

### 実装手順

1. TCP header の data offset から header 長を読む
2. checksum を pseudo header 込みで検証する
3. LISTEN で SYN を受ける
4. `remoteSeq = clientSeq + 1`
5. SYN-ACK を返す
6. final ACK を受けたら ESTABLISHED にする

### Packet の中身

TCP header は最低 20 byte です。options があると長くなるので、payload の開始位置は固定で 20 と決め打ちしません。`DataOffset * 4` で求めます。

| offset | Go slice | byte数 | フィールド | 意味 |
| ---: | --- | ---: | --- | --- |
| 0 | `b[0:2]` | 2 | SrcPort | 送信元 port |
| 2 | `b[2:4]` | 2 | DstPort | 宛先 port。echo は `8080` |
| 4 | `b[4:8]` | 4 | Seq | この segment の先頭 byte 番号 |
| 8 | `b[8:12]` | 4 | Ack | 次に受け取りたい byte 番号 |
| 12 | `b[12] >> 4` | 4bit | DataOffset | TCP header 長。4 byte 単位 |
| 13 | `b[13]` | 1 | Flags | SYN/ACK/PSH/FIN など |
| 14 | `b[14:16]` | 2 | Window | 受信 window |
| 16 | `b[16:18]` | 2 | Checksum | pseudo header 込み |
| 18 | `b[18:20]` | 2 | Urgent | この教材では使わない |
| 可変 | `b[headerLen:]` | 可変 | Payload | echo する byte stream |

Flags は bit field です。

| flag | 値 | 意味 |
| --- | ---: | --- |
| FIN | `0x01` | 接続終了したい |
| SYN | `0x02` | 接続開始したい |
| PSH | `0x08` | data をすぐ上位へ渡してほしい |
| ACK | `0x10` | Ack field が有効 |

TCP で最初に混乱しやすい点は、SYN と FIN は payload がなくても sequence number を 1 消費することです。

### 自動チェック

```sh
scripts/check-chapter.sh 07
```

### 実機確認

```sh
nc 192.168.100.2 8080
```

この章では接続成立まででよいです。

## 第8章 TCP echo

### 目的

ESTABLISHED 後の TCP payload をそのまま返します。

### 実装する場所

- `internal/stack/stack.go`

### 実装手順

1. ACK-only packet には何も返さない
2. payload がある packet だけ処理する
3. `seg.Seq == remoteSeq` を確認する
4. `remoteSeq += len(payload)`
5. PSH/ACK で同じ payload を返す

### 自動チェック

```sh
scripts/check-chapter.sh 08
```

### 実機確認

```sh
nc 192.168.100.2 8080
hello
```

`hello` が返れば合格です。

## 第9章 TCP close

### 目的

`nc` を終了したときに FIN を処理し、次の接続を受けられる状態に戻します。

### 実装する場所

- `internal/stack/stack.go`

### 実装手順

1. FIN を受けたら `remoteSeq = seg.Seq + 1`
2. ACK/FIN を返す
3. 相手の ACK を受けたら LISTEN に戻す
4. 次の `nc` 接続が成功することを確認する

### 自動チェック

```sh
scripts/check-chapter.sh 09
```

## 第10章 デバッグ

Wireshark filter:

```text
arp
icmp
udp.port == 9000
tcp.port == 8080
```

典型的な失敗:

- wrong MAC: Ethernet Src/Dst の入れ替えミス
- bad checksum: pseudo header、length、checksum field zero 化のミス
- wrong ack: SYN/FIN が 1 sequence 消費することの忘れ
- no reply: DstIP、DstPort、EtherType の分岐ミス
