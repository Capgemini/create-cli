package vault

import (
	"time"
)

func Healthcheck() bool {
	logger.Waitingf("Calling Health endpoint...")

	for {
		vaultHealthResponse, err := vaultClient.Sys().Health()
		if err != nil {
			logger.Warningf("Vault still unhealthy, waiting for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		if vaultHealthResponse.Initialized {
			if vaultHealthResponse.Sealed {
				// Due to the fact that we cannot unseal vault as we do not have the unseal keys,
				// we tell user to unseal manually because at this point, Vault is initialised
				logger.Warningf("Vault is initialized and sealed. Please unseal with the unseal keys retrieved on initialization")
				return true
			}

			logger.Warningf("Vault is initialized and unsealed.")
			return true
		}

		// it is assumed that at this point vault is uninitialized
		logger.Actionf("Vault is uninitialized...")
		return false
	}
}
