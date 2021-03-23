@local
Feature: Bootstrap a GitOps repository

    As a user I want to bootstrap a GitOps repository
    and configure my Argo CD applications and OpenShift pipelines.

    # Scenario: Execute KAM bootstrap command with default and --push-to-git=true flag
    #     When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --git-host-access-token $GITHUB_TOKEN --push-to-git=true" succeeds
    #     Then stderr should be empty
    
    Scenario: Execute KAM bootstrap command and check if Argo CD applications are healthy and in-sync
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --git-host-access-token $GITHUB_TOKEN --push-to-git=true" succeeds
        Then stderr should be empty
        And Make gitops repository public
        And Apply Argo CD applications at "gitops/config/argocd" to the cluster
        And Wait for Applications to Sync
        And Argo CD applications "argo-app,cicd-app,dev-app-taxi,dev-env,stage-env" are healthy and in-sync
