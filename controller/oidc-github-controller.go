package controllers // oidc-github-controller reconciles a ManagedCluster object

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/client-go/kubernetes"
	"github.com/google/go-github/v69/github"

)

const GitHubLabelKey = "oidc.open-cluster-management.io/github-enable"

type OIDCGithubManagedClusterReconciler struct {
	client.Client
	Kubeset  kubernetes.Interface
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *OIDCGithubManagedClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("managedCluster", req.NamespacedName)

	log.Info("Reconcile Start.")

	var managedCluster managedclusterv1.ManagedCluster
	if err := r.Get(ctx, req.NamespacedName, &managedCluster); err != nil {
		log.Info("Resource deleted")
		return ctrl.Result{}, nil
	}

	// Now validate the GitHub label is present
	mcLabels := managedCluster.GetLabels()
	enable, found := mcLabels[GitHubLabelKey]
	if !found || enable != "true" {
		// Nothing to do
		log.Info("Nothing to do.")
		return ctrl.Result{}, nil
	}

	// COMBINED TASK, needed for BOTH
	// ✅ 1. Get the Github Secret
	var githubAccess corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: "openshift-config", Name: "create-oauth-github"}, &githubAccess); err != nil {
		log.Info("No GitHub OAuth App secret found.")
		return ctrl.Result{}, nil
	}

	// REMOVE TASKs
	// When deleting the Managed Cluster:
	// 1. remove the GitHub OAuth APP entry for the cluster being deleted
	// ENABLE TASKs
	// ➡️ 1. Pull up relevant cluster information API and OAUTH
	consoleUrl := managedCluster.Spec.console_url
	oauthUrl := strings.Replace(apiUrl, "/api.", "/oauth.apps.")

	//     3. Build a GitHub API connection with the token
	client := github.NewClient(nil).WithAuthToken(githubAccess.Data["github_oauth_app_token"])
	
	//     4. Create an OAuth APP for the cluster, using the included URL and Secret data
    /* It doesn't look like this is supported, without using a browser interaction automation */

	//     5. Get the new OAuth token from GitHub
	//     6. Create a ManifestWork with an authConfig object and secret that will join the cluster to GitHub auth

	log.Info("Reconcile End.  ")
	return ctrl.Result{}, nil
}

func (r *OIDCGithubManagedClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managedclusterv1.ManagedCluster{}).
		Complete(r)
}
