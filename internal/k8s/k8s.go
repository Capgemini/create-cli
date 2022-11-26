package k8s

import (
	"context"
	"create-cli/internal/log"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var logger = log.StderrLogger{Stderr: os.Stderr, Tool: "Kubernetes"}

var defaultSecretName = "create-secrets"
var defaultSecretNamespace = "default"

var k8sClient *kubernetes.Clientset

func InitKubeClient() {
	if k8sClient != nil {
		return
	}

	outCluster := viper.GetBool("out-cluster")
	if outCluster {
		logger.Waitingf("Initialising Out-Cluster K8s Client...")
		initOutClusterKubeClient()
		logger.Successf("K8s Client initialised")
		return
	}
	logger.Waitingf("Initialising In-Cluster K8s Client...")
	initInClusterKubeClient()
	logger.Successf("K8s Client initialised")
}

func initInClusterKubeClient() {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Failuref(`There was an error initialising the in-cluster config. If you want to run create-cli from outside the cluster, be sure to provide the --out-cluster flag.`)
		panic(err.Error())
	}

	k8sClient, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Failuref(`There was an error initialising the in-cluster config. If you want to run create-cli from outside the cluster, be sure to provide the --out-cluster flag.`)
		panic(err.Error())
	}
}

func initOutClusterKubeClient() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config/tools_config"), "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "~/.kube/config/tools_cluster")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

// CreateDefaultOpaqueSecret creates a secret called `secrets` inside the default namespace that
// will hold all of the immediated secrets that an admin will find useful for maintainence of CREATE.
// Any other secrets that are needed will have to be retrieved by going into the relevant namespaces
func CreateDefaultOpaqueSecret() {
	secret := &coreV1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultSecretName,
			Namespace: defaultSecretNamespace,
		},
		Type: "Opaque",
		// here we just put a dummy value in so the data isn't completely empty
		// otherwise when we patch the secret, it will error as there is no data to patch
		Data: map[string][]byte{"Ignore": []byte("Me")},
	}

	logger.Waitingf("Creating secret: %s in namespace: %s...", defaultSecretName, defaultSecretNamespace)
	_, err := k8sClient.CoreV1().Secrets(defaultSecretNamespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		logger.Failuref("Error creating secret: %s in namespace: %s", defaultSecretName, defaultSecretNamespace)
		panic(err)
	}

	logger.Successf("Secret created.")

}

type PatchSecretRequest struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func doesDefaultOpaqueSecretExist() bool {
	_, err := k8sClient.CoreV1().Secrets(defaultSecretNamespace).Get(context.Background(), defaultSecretName, metav1.GetOptions{})
	if err != nil {
		logger.Warningf("Error retrieving secret: %s in namespace: %s, this means it probably doesn't exist", defaultSecretName, defaultSecretNamespace)
		return false
	}
	return true
}

func PatchOpaqueSecret(patch []PatchSecretRequest) {
	secretExist := doesDefaultOpaqueSecretExist()
	if !secretExist {
		CreateDefaultOpaqueSecret()
	}

	logger.Waitingf("Patching secret..")
	patchBytes, _ := json.Marshal(patch)
	_, err := k8sClient.CoreV1().Secrets(defaultSecretNamespace).Patch(context.Background(), defaultSecretName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})

	if err != nil {
		logger.Failuref("Error patching secret: %s in namespace: %s", defaultSecretName, defaultSecretNamespace)
		panic(err)
	}

	logger.Successf("Secret patched")
}

func GetSecret(secretName string, namespace string) *v1.Secret {
	logger.Waitingf("Retrieving secret: %s...", secretName)
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		logger.Failuref("Error retrieving secret %s", secretName)
		panic(err.Error())
	}
	logger.Successf("Retrieved secret: %s", secretName)
	return secret
}
