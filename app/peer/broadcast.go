package peer

import (
	"bufio"
	"fmt"
	"myapp/app/pkg/signature"
	"net"

	"github.com/bytedance/sonic"
)

// BroadcastBlockchain mengirimkan seluruh blockchain ke semua peers yang terhubung.
func (p2p *P2PNetwork) BroadcastBlockchain() {
	blockchainData := p2p.Blockchain.GetAllBlocks(p2p.dbConn)

	fmt.Println("blockchainData:", blockchainData)

	inventory := Inv{Data: blockchainData}
	payload := GobEncode(inventory)

	r, s, err := signature.SignData(p2p.privateKey, string(payload))
	if err != nil {
		fmt.Println("Failed to sign blockchain data:", err)
		return
	}
	message := Message{
		Type: BlockchainUpdate,
		Data: payload,
		Signature: signature.Signature{
			R: r,
			S: s,
		},
		SenderAddress: p2p.localAddr,
	}
	data, err := sonic.Marshal(message)
	if err != nil {
		fmt.Println("gagal membuat request JSON:", err)
		return
	}
	data = append(data, '\n')

	for _, peer := range p2p.peers {
		fmt.Println("broadcasting blockchain to peer:", peer)
		conn, err := net.Dial("tcp", peer.Address)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			continue
		}
		defer conn.Close()

		writer := bufio.NewWriter(conn)
		_, err = writer.WriteString(string(data))
		if err != nil {
			fmt.Println("gagal mengirim request:", err)
			continue
		}
		writer.Flush()
	}
}
