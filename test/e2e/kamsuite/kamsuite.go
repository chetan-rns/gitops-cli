package kamsuite

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/redhat-developer/kam/pkg/pipelines/git"
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// KAM related steps
	s.Step(`^directory "([^"]*)" should exist$`,
		DirectoryShouldExist)

	s.Step(`^Apply Argo CD applications at "([^"]*)" to the cluster`, kubectlKustomizeApply)

	s.Step(`^Argo CD applications "([^"]*)" are healthy and in-sync`, argocdAppStatus)

	s.Step("^Wait", wait)

	s.Step("^Make gitops repository public", makeRepoPublic)

	s.BeforeSuite(func() {
		fmt.Println("Before suite")
		if !envVariableCheck() {
			os.Exit(1)
		}
	})

	s.AfterSuite(func() {
		fmt.Println("After suite")
		// Checking it for local test
		_, ci := os.LookupEnv("CI")
		if !ci {
			deleteGhRepoStep1 := []string{"alias", "set", "repo-delete", `api -X DELETE "repos/$1"`}
			deleteGhRepoStep2 := []string{"repo-delete", strings.Split(strings.Split(os.Getenv("GITOPS_REPO_URL"), "github.com/")[1], ".")[0]}
			ok, _ := executeGhRepoDeleteCommad(deleteGhRepoStep1)
			if !ok {
				os.Exit(1)
			}
			ok, errMessage := executeGhRepoDeleteCommad(deleteGhRepoStep2)
			if !ok {
				fmt.Println(errMessage)
			}
		}
	})

	s.BeforeFeature(func(this *messages.GherkinDocument) {
		fmt.Println("Before feature")
	})

	s.AfterFeature(func(this *messages.GherkinDocument) {
		fmt.Println("After feature")
	})
}

func envVariableCheck() bool {
	envVars := []string{"SERVICE_REPO_URL", "GITOPS_REPO_URL", "IMAGE_REPO", "DOCKERCONFIGJSON_PATH", "GITHUB_TOKEN"}
	val, ok := os.LookupEnv("CI")
	if !ok {
		for _, envVar := range envVars {
			_, ok := os.LookupEnv(envVar)
			if !ok {
				fmt.Printf("%s is not set\n", envVar)
				return false
			}
		}
	} else {
		if val == "prow" {
			fmt.Printf("Running e2e test in OpenShift CI\n")
			os.Setenv("SERVICE_REPO_URL", "https://github.com/kam-bot/taxi")
			os.Setenv("GITOPS_REPO_URL", "https://github.com/kam-bot/taxi-"+os.Getenv("PRNO"))
			os.Setenv("IMAGE_REPO", "quay.io/kam-bot/taxi")
			os.Setenv("DOCKERCONFIGJSON_PATH", os.Getenv("KAM_QUAY_DOCKER_CONF_SECRET_FILE"))
		} else {
			fmt.Printf("You cannot run e2e test locally against OpenShift CI\n")
			return false
		}
		return true
	}
	return true
}

func executeGhRepoDeleteCommad(arg []string) (bool, string) {
	var stderr bytes.Buffer
	cmd := exec.Command("gh", arg...)
	fmt.Println("gh command is : ", cmd.Args)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false, stderr.String()
	}
	return true, stderr.String()
}

func wait() error {
	duration := 1 * time.Minute
	time.Sleep(duration)
	return nil
}

func makeRepoPublic() error {
	repoName, err := repoFromURL(os.Getenv("GITOPS_REPO_URL"))
	fmt.Println(repoName)
	if err != nil {
		return err
	}
	args := []string{"api", "repos/" + repoName, "-F", "private=false"}
	cmd := exec.Command("gh", args...)
	fmt.Println("GH Command : ", cmd.Args)
	return cmd.Run()
}

func repoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return git.GetRepoName(u)
}
