# go-tcp-ip

Go と Linux TAP デバイスで TCP/IP プロトコルスタックを自作するハンズオン教材です。

この `main` branch には完成コードを置いていません。いきなり答えを読むのではなく、章ごとに小さく実装していくための入口です。

完成版は `final` branch にあります。

```sh
git fetch origin
git checkout final
```

## 進め方

1. この branch で [docs/HANDS_ON.md](docs/HANDS_ON.md) を読む
2. 自分で `tcpip-go/` プロジェクトを作る
3. 各章の TODO を実装する
4. 詰まったら `final` branch の同名 package を参照する
5. `go test ./...` と Linux TAP 上の動作確認で進める

## 最終ゴール

Linux host から TAP デバイス越しに、以下が動くことを目標にします。

```sh
ping 192.168.100.2
nc -u 192.168.100.2 9000
nc 192.168.100.2 8080
```

## なぜ main に完成コードを置かないか

この教材の目的は「TCP/IP を packet level で理解すること」です。完成コードを最初から置くと、読者は写経ではなく眺めるだけになりがちです。

そのため:

- `main`: ハンズオンの道筋
- `final`: 完成リファレンス実装

に分けています。

## 必要環境

- Linux
- Go 1.23 以上
- `iproute2`
- `ping`, `arping`, `nc`
- 任意: Wireshark または tcpdump

TAP デバイスを使う実行確認は Linux 前提です。macOS / Windows では Linux VM を使ってください。

## TAP 作成

```sh
sudo ip tuntap add dev tap0 mode tap user "$USER"
sudo ip addr add 192.168.100.1/24 dev tap0
sudo ip link set tap0 up
```

自作スタック側は以下に固定します。

- MAC: `02:00:00:00:00:02`
- IPv4: `192.168.100.2`
- UDP echo: `9000`
- TCP echo: `8080`

## 参考 RFC

- RFC 826 ARP
- RFC 791 IPv4
- RFC 792 ICMP
- RFC 768 UDP
- RFC 793 TCP

