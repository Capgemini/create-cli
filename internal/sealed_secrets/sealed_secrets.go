package sealed_secrets

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"html/template"
	"log"

	// "go.mozilla.org/sops/v3/cmd/groups"

	"filippo.io/age"
	"golang.org/x/crypto/ssh"
)

type Secret struct {
	Metadata Metadata `yaml:"metadata"`
	Data     Data     `yaml:"data"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Data struct {
	TLSCert string `yaml:"tls.crt"`
	TLSKey  string `yaml:"tls.key"`
}

var temp *template.Template

func SealedSecrets() {

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		fmt.Println(err)
		return
	}

	pubKey := identity.Recipient().String()
	log.Println(pubKey)
	// privKey := identity.String()

	// cmd = exec.Command("sops", "-age=age1ykphy2fuc0rmtewtpml69670s9dydkneum384tsj6z480lljmvqqx4kj8u", "--encrypt", "--encrypted-regex", "'^(data|stringData)$'", "--in-place", "internal/sealed_secrets/sealed_secrets.yaml")
	// stdout, err = cmd.Output()

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Print(string(stdout))

	//////////////////////////////
	// temp = template.Must(template.ParseFiles("internal/sealed_secrets/sealed_secrets_template.yaml"))

	// config := Secret{
	// 	Metadata: Metadata{
	// 		Name: "sealed-secrets-keys",
	// 	},
	// 	Data: Data{
	// 		TLSCert: base64.StdEncoding.EncodeToString([]byte(pubKey)),
	// 		TLSKey:  base64.StdEncoding.EncodeToString([]byte(privKey)),
	// 	},
	// }

	// f, err := os.Create("internal/sealed_secrets/sealed_secrets.yaml")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// err = temp.Execute(f, config)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

}

// GenerateECDSAKeys generates ECDSA public and private key pair with given size for SSH.
func generateECDSAKeys() (pubKey string, privKey string) {
	// generate private key
	// logger.Waitingf("Generating SSH Keys for Concourse...")
	var privateKey *ecdsa.PrivateKey
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		// logger.Failuref("Error generating SSH Keys for Concourse", err)
		panic(err)
	}

	var publicKey ssh.PublicKey
	publicKey, err = ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		// logger.Failuref("Error creating public SSH Key for Concourse", err)
		panic(err)
	}
	pubBytes := ssh.MarshalAuthorizedKey(publicKey)

	// encode private key
	var bytes []byte
	bytes, err = x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		// logger.Failuref("Error marshalling private SSH Key for Concourse", err)
		panic(err)
	}
	privBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: bytes,
	})

	// logger.Successf("Generated SSH Keys for Concourse")
	return string(pubBytes), string(privBytes)
}
