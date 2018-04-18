package plasma

import (
	"crypto/ecdsa"

	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

// TODO: move to client with sub args for deposit args.
func DepositIntegrationTest(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")

	var privateKeyECDSA *ecdsa.PrivateKey

	if exists(userAddress) && exists(privateKey) {
		privateKeyECDSA = util.ToPrivateKeyECDSA(privateKey)
	} else if exists(keystoreDir) &&
		exists(keystoreFile) &&
		exists(userAddress) {
		keyWrapper := util.GetFromKeyStore(userAddress, keystoreDir, keystoreFile, signPassphrase)
		privateKeyECDSA = keyWrapper.PrivateKey
	}

	if privateKeyECDSA == nil {
		panic("Private key ecdsa not found")
	}

	plasma := CreatePlasmaClient(nodeURL, contractAddress)

	depositValue := 1000000000

	t := createDepositTx(userAddress, depositValue)
	Deposit(plasma, privateKeyECDSA, userAddress, 1000000000, &t)
}
