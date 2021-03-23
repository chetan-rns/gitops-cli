package kamsuite

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultArgoCDInstanceNamespace = "openshift-gitops"
)

func kubectlKustomizeApply(path string) error {
	args := []string{"apply", "-k", path}
	return kubectlCommand(args...)
}

func argocdAppStatus(appString string) error {
	apps := strings.Split(appString, ",")
	fmt.Println(apps)
	for _, app := range apps {
		fmt.Printf("checking if application %q is healthy and in-sync", app)

		ns := types.NamespacedName{
			Name:      app,
			Namespace: defaultArgoCDInstanceNamespace,
		}

		healthStatus, err := getHealthStatus(ns)
		if err != nil {
			return err
		}
		fmt.Println(healthStatus)
		if healthStatus != "'Healthy'" {
			return fmt.Errorf("application %q is unhealthy", app)
		}

		syncStatus, err := getSyncStatus(ns)
		if err != nil {
			return err
		}
		if syncStatus != "'Synced'" {
			return fmt.Errorf("application %q is out of sync", app)
		}

		fmt.Printf("application %q is healthy and in-sync", app)
	}

	return nil
}

func getHealthStatus(ns types.NamespacedName) (string, error) {
	args := []string{"get", "app", ns.Name, "-n", ns.Namespace, "-ojsonpath='{.status.health.status}'"}
	b, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getSyncStatus(ns types.NamespacedName) (string, error) {
	args := []string{"get", "app", ns.Name, "-n", ns.Namespace, "-ojsonpath='{.status.sync.status}'"}
	b, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readFromStdIn() []byte {
	var resource []byte
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.Read(resource)
	exitOnError(err)
	return resource
}

func kubectlCommand(args ...string) error {
	return exec.Command("kubectl", args...).Run()
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getArgoCDClusterSecret() {

	args := []string{"get", "secret", "argocd-cluster-cluster", "-n", defaultArgoCDInstanceNamespace, `-ojsonpath='{.data.admin\.password}'`, "|", "base64", "-d"}

}
